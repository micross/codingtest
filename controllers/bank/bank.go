package bank

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/micross/codingtest/controllers/common"
	"github.com/micross/codingtest/models"
	"github.com/micross/codingtest/utils"
	"github.com/rs/xid"
)

func Withdraw(c *gin.Context) {
	type ReqData struct {
		AccountID string `json:"account_id" binding:"required"`
		Amount    string `json:"amount" binding:"gt=0,required"`
	}

	var reqData ReqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		common.WriteError(err.Error(), c)
		return
	}

	var account models.Account
	if err := models.DB.Where("id = ?", reqData.AccountID).Where("status = ?", 1).First(&account).Error; err != nil {
		common.WriteError("account not exists", c)
		return
	}

	tmp, _ := strconv.ParseInt(reqData.Amount, 10, 64)
	amount := tmp * 100

	if account.Balance < amount {
		common.WriteError("insufficient account balance", c)
		return
	}

	account.Balance = account.Balance - amount

	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		common.WriteError(tx.Error.Error(), c)
		return
	}

	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		common.WriteError(err.Error(), c)
		return
	}

	journal := models.Journal{
		ID:            xid.New().String(),
		FromAccountID: reqData.AccountID,
		ToAccountID:   reqData.AccountID,
		Amount:        -amount,
		Charge:        0,
		Status:        1,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(&journal).Error; err != nil {
		tx.Rollback()
		common.WriteError(err.Error(), c)
		return
	}

	if err := tx.Commit().Error; err != nil {
		common.WriteError(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func Deposit(c *gin.Context) {
	type ReqData struct {
		AccountID string `json:"account_id" binding:"required"`
		Amount    string `json:"amount" binding:"gt=0,required"`
	}

	var reqData ReqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		common.WriteError(err.Error(), c)
		return
	}

	var account models.Account
	if err := models.DB.Where("id = ?", reqData.AccountID).Where("status = ?", 1).First(&account).Error; err != nil {
		common.WriteError("account not exists", c)
		return
	}

	tmp, _ := strconv.ParseInt(reqData.Amount, 10, 64)
	amount := tmp * 100

	account.Balance = account.Balance + amount

	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		common.WriteError(tx.Error.Error(), c)
		return
	}

	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		common.WriteError(err.Error(), c)
		return
	}

	journal := models.Journal{
		ID:            xid.New().String(),
		FromAccountID: reqData.AccountID,
		ToAccountID:   reqData.AccountID,
		Amount:        amount,
		Charge:        0,
		Status:        1,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(&journal).Error; err != nil {
		tx.Rollback()
		common.WriteError(err.Error(), c)
		return
	}

	if err := tx.Commit().Error; err != nil {
		common.WriteError(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func Transfer(c *gin.Context) {
	type ReqData struct {
		FromAccountID string `json:"from_account_id" binding:"required"`
		ToAccountID   string `json:"to_account_id" binding:"required"`
		Amount        string `json:"amount" binding:"gt=0,required"`
		Name          string `json:"name" binding:"required"`
	}

	var reqData ReqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		common.WriteError(err.Error(), c)
		return
	}

	RedisConn := models.RedisPool.Get()
	defer RedisConn.Close()

	used, err := redis.Int64(RedisConn.Do("GET", reqData.FromAccountID))
	if err != nil {
		fmt.Println("redis get failed:", err)
	}

	tmp, _ := strconv.ParseInt(reqData.Amount, 10, 64)
	amount := tmp * 100

	if (used + amount) > (10000 * 100) {
		common.WriteError("reached the maximum transfer limit", c)
		return
	}

	RedisConn.Do("INCRBY", reqData.FromAccountID, amount)
	exptime := utils.TomorrowUnix()
	RedisConn.Do("PEXPIREAT", reqData.FromAccountID, exptime)

	var fromAccount models.Account
	if err := models.DB.Where("id = ?", reqData.FromAccountID).Where("status = ?", 1).First(&fromAccount).Error; err != nil {
		common.WriteError("from account not exists", c)
		return
	}

	var fromOwner models.Owner
	if err := models.DB.Where("id = ?", fromAccount.OwnerId).First(&fromOwner).Error; err != nil {
		common.WriteError("from owner not exists", c)
		return
	}

	var serviceCharge int64 = 0
	flag := 1

	if fromOwner.Name != reqData.Name {
		serviceCharge = 100 * 100
		flag = 2
	}

	if fromAccount.Balance < (amount + serviceCharge) {
		common.WriteError("insufficient account balance", c)
		return
	}

	fromAccount.Balance = fromAccount.Balance - amount - serviceCharge

	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		common.WriteError(tx.Error.Error(), c)
		return
	}

	if err := tx.Save(&fromAccount).Error; err != nil {
		tx.Rollback()
		common.WriteError(err.Error(), c)
		return
	}

	journalID := xid.New().String()
	journal := models.Journal{
		ID:            journalID,
		FromAccountID: reqData.FromAccountID,
		ToAccountID:   reqData.ToAccountID,
		Amount:        amount,
		Charge:        serviceCharge,
		Status:        2,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(&journal).Error; err != nil {
		tx.Rollback()
		common.WriteError(err.Error(), c)
		return
	}

	if err := tx.Commit().Error; err != nil {
		common.WriteError(err.Error(), c)
		return
	}

	msg := fmt.Sprintf("%s#%s#%s#%s#%d", reqData.FromAccountID, reqData.ToAccountID, reqData.Name, journalID, flag)
	models.SendToMQ(msg)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
