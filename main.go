package main

import (
	"IM/models"
	"IM/router"
	"IM/utils"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"time"
)

// init
//
//	@Description: 进行一系列初始化任务
func init() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()
	models.InitUDPProc()
	InitTimer()
}

// main
//
//	@Description: 主函数
func main() {
	r := router.Router()
	r.Run(viper.GetString("port.server"))
}

// InitTimer
//
//	@Description: 初始化定时器
func InitTimer() {
	utils.Timer(time.Duration(viper.GetInt("timeout.DelayHeartbeat"))*time.Second, time.Duration(viper.GetInt("timeout.HeartbeatHz"))*time.Second, models.CleanConnection, "")
	color.Green("Init Timer Successfully!")
}
