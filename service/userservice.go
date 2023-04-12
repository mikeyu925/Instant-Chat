package service

import (
	"IM/models"
	"IM/utils"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GetUserList
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [post]
func GetUserList(c *gin.Context) {
	list := models.GetUserList()
	c.JSON(200, gin.H{
		"message": list,
	})
}

// CreateUser
// @Tags 用户模块
// @Summary 新增用户
// @param name query string false "用户名"
// @param password query string false "密码"
// @param Identity query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("Identity")
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户名或密码不能为空！",
			"data":    user,
		})
		return
	}
	// 查看用户名是否被注册
	data := models.FindUserByName(user.Name)
	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户名已注册！",
			"data":    user,
		})
		return
	}
	// 验证两次密码是否相同
	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "两次密码不一致！",
			"data":    user,
		})
		return
	}

	salt := fmt.Sprintf("%06d", rand.Int31()) // 生成随机数
	// md5加密
	user.PassWord = utils.MakePassword(password, salt) // 存储的是加密后的字符串
	user.Salt = salt                                   // 生成的随机数
	// 设置时间
	nowTime := time.Now()
	user.LoginTime = nowTime
	user.LoginOutTime = nowTime
	user.HeartbeatTime = nowTime
	// 进行用户注册插入Mysql
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "新增用户成功！",
		"data":    user,
	})
}

// DeleteUser
// @Tags 用户模块
// @Summary 删除用户「只做逻辑删除」
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [post]
func DeleteUser(c *gin.Context) {
	// 先判断用户是否存在
	idStr := c.Query("id")
	data := models.FindUserById(idStr)
	if data.Name == "" { //
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户不存在！",
		})
		return
	}
	user := models.UserBasic{}
	Id, _ := strconv.Atoi(idStr)
	user.ID = uint(Id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"message": "delete user success!",
	})
}

// UpdateUser
// @Tags 用户模块
// @Summary 修改用户
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	// 读取修改信息
	user := models.UserBasic{}
	idStr := c.PostForm("id")
	id, _ := strconv.Atoi(idStr)
	// 先判断用户是否存在
	data := models.FindUserById(idStr)
	if data.Name == "" { //
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户不存在！",
		})
		return
	}
	// 判断更改的昵称是否存在
	newName := c.PostForm("name")
	checkUser := models.FindUserByName(newName)
	if checkUser.ID != uint(id) {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "您想要更改的用户名已经存在！",
		})
		return
	}
	// 获得新、旧密码
	oldpwd := c.PostForm("oldpwd")
	newpwd := c.PostForm("newpwd")
	// 校验旧密码是否正确
	md5OldPwd := utils.MakePassword(oldpwd, data.Salt) // 获得加密后的字符串
	if strings.Compare(md5OldPwd, data.PassWord) != 0 {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "旧密码错误",
		})
		return
	}

	user.ID = uint(id)
	user.Name = c.PostForm("name")
	// 生成新密码
	user.PassWord = utils.MakePassword(newpwd, data.Salt) // 获得加密后的字符串
	user.Phone = c.PostForm("phone")
	user.Avatar = c.PostForm("icon")
	user.Email = c.PostForm("email")

	// 进行格式验证
	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "修改格式不匹配！",
			"data":    user,
		})
	}
	// 更新用户密码
	models.UpdateUser(user)

	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "修改用户成功！",
		"data":    user,
	})

}

// FindUserByNameAndPwd
// @Summary 登陆
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")

	// 通过名字查找用户
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "该用户不存在",
			"data":    data,
		})
		return
	}
	// 验证密码
	flag := utils.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "密码不正确",
			"data":    data,
		})
		return
	}
	// 查询用户
	pwd := utils.MakePassword(password, user.Salt) // 通过用户名+加密的密码再数据库中查找用户信息
	data = models.FindUserByNameAndPwd(name, pwd)

	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "登录成功",
		"data":    data,
	})
}

// 防止跨域站点伪造请求
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SearchFriends
//
//	@Description: 查找用户userId的好友
//	@param c
func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId")) // 获取用户id
	users := models.SearchFriend(uint(id))               // 得到用户的好友信息
	utils.RespOKList(c.Writer, users, len(users))
}

// AddFriend
//
//	@Description: 添加好友handler函数
//	@param c
func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	code, msg := models.AddFriend(uint(userId), targetName)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

//
// DeleteFriend
//  @Description: 删除好友
//  @param c
//
func DeleteFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	code, msg := models.DeleteFriend(uint(userId), targetName)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// CreateCommunity
//
//	@Description: 创建群聊handler
//	@param c
func CreateCommunity(c *gin.Context) {
	// 获取要创建的群信息
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	name := c.Request.FormValue("name")
	icon := c.Request.FormValue("icon")
	desc := c.Request.FormValue("desc")
	community := models.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Img = icon
	community.Desc = desc
	// 创建群
	code, msg := models.CreateCommunity(community)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// LoadCommunity
//
//	@Description: 加载群列表
//	@param c
func LoadCommunity(c *gin.Context) {
	// 获取用户id
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	data, msg := models.LoadCommunity(uint(ownerId))
	if len(data) != 0 {
		utils.RespList(c.Writer, 0, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// JoinGroups
//
//	@Description: 加入群
//	@param c
func JoinGroups(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId")) // 获取申请加入者id
	comId := c.Request.FormValue("comId")                    // 群id
	data, msg := models.JoinGroup(uint(userId), comId)
	if data == 0 {
		utils.RespOK(c.Writer, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}
func FindByID(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	data := models.FindByID(uint(userId))
	if data.Name == "" {
		utils.RespFail(c.Writer, "not find by id")
	}
	utils.RespOK(c.Writer, data, "ok")
}

// RedisMsg
//
//	@Description: 查找两个用户A和用户B的聊天记录缓存
//	@param c
func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	res := models.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	utils.RespOKList(c.Writer, "ok", res)
}
