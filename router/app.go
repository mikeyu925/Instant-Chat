package router

import (
	"IM/docs"
	"IM/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()
	// 加载静态资源
	r.Static("/asset", "asset/")
	r.StaticFile("/favicon.ico", "asset/images/favicon.ico")
	r.LoadHTMLGlob("views/**/*")
	//r.StaticFS()

	// swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	// 首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.POST("/searchFriends", service.SearchFriends)

	// 用户模块
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	r.POST("/user/find", service.FindByID)
	//发送消息
	r.GET("/user/sendMsg", service.SendMsg)
	// 缓存
	r.POST("/user/redisMsg", service.RedisMsg)

	//添加好友
	r.POST("/contact/addfriend", service.AddFriend)
	//创建群聊
	r.POST("/contact/createCommunity", service.CreateCommunity)
	// 加载群列表
	r.POST("/contact/loadcommunity", service.LoadCommunity)
	// 加入群组
	r.POST("/contact/joinGroup", service.JoinGroups)
	//上传文件
	r.POST("/attach/upload", service.Upload)

	return r
}
