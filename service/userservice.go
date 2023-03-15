package service

import (
	"IM/models"
	"IM/utils"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
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
	//fmt.Println(user.Name, "  >>>>>>>>>>>  ", password, repassword)

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
	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "两次密码不一致！",
			"data":    user,
		})
		return
	}
	// 生成随机数
	salt := fmt.Sprintf("%06d", rand.Int31())
	// md5加密
	user.PassWord = utils.MakePassword(password, salt) // 存储的是加密后的字符串
	user.Salt = salt
	//fmt.Println(user.PassWord)

	nowTime := time.Now()
	user.LoginTime = nowTime
	user.LoginOutTime = nowTime
	user.HeartbeatTime = nowTime

	// 进行用户注册
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
	// 先判断用户是否存在
	data := models.FindUserById(idStr)
	if data.Name == "" { //
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户不存在！",
		})
		return
	}

	id, _ := strconv.Atoi(idStr)
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	pwd := c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Avatar = c.PostForm("icon")
	user.Email = c.PostForm("email")
	fmt.Println("update :", user)
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
	// 这里salt是个空的！！！！ 需要先查出来 salt
	user.PassWord = utils.MakePassword(pwd, data.Salt)
	models.UpdateUser(user)

	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "修改用户成功！",
		"data":    user,
	})

}

// FindUserByNameAndPwd
// @Summary 所有用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")

	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "该用户不存在",
			"data":    data,
		})
		return
	}

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

func FindByID(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))

	//	name := c.Request.FormValue("name")
	data := models.FindByID(uint(userId))
	utils.RespOK(c.Writer, data, "ok")
}
