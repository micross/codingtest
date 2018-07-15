package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micross/codingtest/controllers/common"
	"github.com/micross/codingtest/models"
	"github.com/rs/xid"
)

func Balance(c *gin.Context) {
	accountID := c.Param("id")
	var account models.Account
	if err := models.DB.Where("id = ?", accountID).Where("status = ?", 1).First(&account).Error; err != nil {
		common.WriteError("account not exists", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": account.Balance / 100,
	})
}

func Open(c *gin.Context) {
	var account models.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		common.WriteError("param error", c)
		return
	}

	id := xid.New().String()
	account.ID = id
	account.Status = 1
	account.Balance = account.Balance * 100

	if account.OwnerId == "" {
		common.WriteError("missing owner id", c)
		return
	}

	var owner models.Owner
	if err := models.DB.Where("id = ?", account.OwnerId).First(&owner).Error; err != nil {
		common.WriteError("owner not exists", c)
		return
	}

	if err := models.DB.Create(&account).Error; err != nil {
		common.WriteError("param error", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func Close(c *gin.Context) {
	accountID := c.Param("id")
	var account models.Account
	if err := models.DB.Where("id = ?", accountID).First(&account).Error; err != nil {
		common.WriteError("account not exists", c)
		return
	}
	account.Status = 2
	if err := models.DB.Save(&account).Error; err != nil {
		common.WriteError("error", c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
