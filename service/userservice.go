package service

import (
	"IM/models"
	"github.com/gin-gonic/gin"
)

// GetUserList
// @Tags 首页
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	list := models.GetUserList()
	c.JSON(200, gin.H{
		"message": list,
	})
}
