package models

import "gorm.io/gorm"

type RelationType int32

const (
	Friend RelationType = 1 // 好友
	Group  RelationType = 2 // 群
)

// 用户关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息
	TargetId uint //对应的谁   /群 ID
	Type     int  //对应的类型  1好友  2群  3xx
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}
