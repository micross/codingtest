package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/micross/codingtest/controllers/account"
	"github.com/micross/codingtest/controllers/bank"
)

// Routes
func Route(r *gin.Engine) {
	r.GET("/accounts/:id", account.Balance)  // get current balance
	r.POST("/accounts", account.Open)        // open account
	r.DELETE("/accounts/:id", account.Close) // close account
	r.POST("/withdraw", bank.Withdraw)       // withdraw
	r.POST("/deposit", bank.Deposit)         // deposit
	r.POST("/transfer", bank.Transfer)       // transfer
}
