package models

import "gorm.io/gorm"

type UserBasic struct {
	gorm.Model    // 自动增加4个字段
	Name          string
	PassWord      string
	Phone         string
	Email         string
	Identity      string
	ClientIP      string
	ClientPort    string
	LoginTime     uint64 // 登陆时间
	HeartbeatTime uint64 // 心跳时间
	LogOutTime    uint64 // 下线时间
	IsLogout      bool
	DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}
