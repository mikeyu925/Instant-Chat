package router

import (
	"IM/docs"
	"IM/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default() // 初始化 gin 引擎
	// 加载静态资源
	{
		r.Static("/asset", "asset/")
		r.StaticFile("/favicon.ico", "asset/images/favicon.ico")
		r.LoadHTMLGlob("views/**/*")
	}

	// swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	r.GET("/", service.GetIndex)             // 首页
	r.GET("/toRegister", service.ToRegister) // 返回注册页面
	r.GET("/toChat", service.ToChat)         // 登陆之后跳转的页面
	//r.GET("/index", service.GetIndex)

	r.GET("/chat", service.Chat)

	// 基础功能组件路由组
	base := r.Group("/base")
	{
		base.POST("/upload", service.Upload) //上传文件
	}

	// 用户基础模块路由组
	userRouter := r.Group("/user")
	{
		userRouter.POST("/getUserList", service.GetUserList)                   // 获取用户列表  --- 暂时无用
		userRouter.POST("/deleteUser", service.DeleteUser)                     // 删除用户  --- 暂时无用
		userRouter.POST("/createUser", service.CreateUser)                     // 创建「注册」用户
		userRouter.POST("/updateUser", service.UpdateUser)                     // 更新用户信息
		userRouter.POST("/findUserByNameAndPwd", service.FindUserByNameAndPwd) // 用户登陆
		userRouter.POST("/find", service.FindByID)                             // 查找用户
		userRouter.GET("/sendMsg", service.SendMsg)                            // 发送消息
		userRouter.POST("/redisMsg", service.RedisMsg)                         // 缓存
		userRouter.POST("/searchFriends", service.SearchFriends)               // 查找好友
	}

	// 用户关系相关路由组
	contactRouter := r.Group("/contact")
	{
		contactRouter.POST("/addfriend", service.AddFriend)             //添加好友
		contactRouter.POST("/createCommunity", service.CreateCommunity) //创建群聊
		contactRouter.POST("/loadcommunity", service.LoadCommunity)     // 加载群列表
		contactRouter.POST("/joinGroup", service.JoinGroups)            // 加入群组
	}

	return r
}
