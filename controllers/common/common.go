package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteError(msg string, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "error",
		"message": msg,
	})
	// stop the request chain
	c.Abort()
}
