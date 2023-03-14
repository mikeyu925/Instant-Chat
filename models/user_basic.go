package models

import (
	"gorm.io/gorm"
	"time"
)

type UserBasic struct {
	gorm.Model
	Name          string
	PassWord      string
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Avatar        string //头像
	Identity      string
	ClientIp      string
	ClientPort    string
	Salt          string
	LoginTime     time.Time
	HeartbeatTime time.Time
	LoginOutTime  time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout      bool
	DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}
