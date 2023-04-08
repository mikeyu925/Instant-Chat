package models

import (
	"IM/utils"
	"fmt"
	"gorm.io/gorm"
)

// Community
// @Description: 群结构
type Community struct {
	gorm.Model
	Name    string // 群名
	OwnerId uint   // 创建者
	Img     string // 群头像
	Desc    string // 群描述
}

// TableName
//
//	@Description: 群表的表名设置函数
//	@receiver table
//	@return string
func (table *Community) TableName() string {
	return "community"
}

// CreateCommunity
//
//	@Description: 创建群聊
//	@param community 群信息
//	@return int 0：成功 -1：失败
//	@return string  响应信息
func CreateCommunity(community Community) (int, string) {
	if len(community.Name) == 0 {
		return -1, "群名称不能为空，请输入"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}
	// 开启事务
	tx := utils.DB.Begin()
	defer func() {
		//事务一旦开始，不论什么异常最终都会 Rollback
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(string(err.Error()))
		tx.Rollback()
		return -1, "建群失败"
	}
	contact := Contact{}
	contact.OwnerId = community.OwnerId
	contact.TargetId = community.ID
	contact.Type = 2 //群关系
	if err := utils.DB.Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "添加群关系失败"
	}
	// 提交事务
	tx.Commit()
	return 0, "建群成功"
}

// LoadCommunity
//
//	@Description: 加载群
//	@param ownerId  当前用户Id
//	@return []*Community  其加入的群信息
//	@return string
func LoadCommunity(ownerId uint) ([]*Community, string) {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type=2", ownerId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}

	data := make([]*Community, 10)
	utils.DB.Where("id in ?", objIds).Find(&data)
	return data, "查询成功"
}
