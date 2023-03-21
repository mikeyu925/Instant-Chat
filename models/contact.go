package models

import (
	"IM/utils"
	"gorm.io/gorm"
)

type RelationType int32

const (
	Friend RelationType = 1 // 好友
	Group  RelationType = 2 // 群
)

// 用户关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息
	TargetId uint //对应的谁 /群 ID
	Type     int  //对应的类型  1好友  2群  3xx
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

// 查找用户好友列表
func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	// 通过Mysql 查询当前用户的好友信息
	utils.DB.Where("owner_id = ? and type = ?", userId, Friend).Find(&contacts)
	//utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0) // 通过得到的好友id，再进行一遍查询，返回好友的信息
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}
