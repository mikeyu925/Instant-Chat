package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello!",
	})
}
