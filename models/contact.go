package models

import (
	"IM/utils"
	"gorm.io/gorm"
)

type RelationType int32

const (
	_FriendType RelationType = 1 // 好友
	_GroupType  RelationType = 2 // 群
)

// 用户关系
type Contact struct {
	gorm.Model
	OwnerId  uint   //谁的关系信息
	TargetId uint   //对应的 用户/群 ID
	Type     int    //对应的类型  1好友  2群  考虑如何实现黑名单
	Desc     string // 描述？
}

func (table *Contact) TableName() string {
	return "contact"
}

// SearchFriend
//
//	@Description: 查找用户好友列表
//	@param userId 当前用户id
//	@return []UserBasic 其好友列表
func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	// 通过Mysql 查询当前用户的好友信息
	utils.DB.Where("owner_id = ? and type = ?", userId, _FriendType).Find(&contacts)
	// 遍历所有关系，获得targetIDs
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0) // 通过得到的好友id，再进行一遍查询，返回好友的信息
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}

// AddFriend
//
//	@Description: 添加好友
//	@param userId 当前用户的id
//	@param targetName  想要添加好友的名字
//	@return int  0：成功 -1：失败
//	@return string 响应信息
func AddFriend(userId uint, targetName string) (int, string) {
	// 查找用户名
	if targetName != "" {
		targetUser := FindUserByName(targetName) // 通过名字查找用户
		if targetUser.Salt != "" {
			if targetUser.ID == userId {
				return -1, "不能添加自己"
			}
			// 判断是否已经是好友关系
			contact0 := Contact{}
			utils.DB.Where("owner_id =?  and target_id =? and type=1", userId, targetUser.ID).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}
			// 开启一个事务
			tx := utils.DB.Begin()
			defer func() { //事务一旦开始，不论什么异常最终都会 Rollback
				// 如果出现异常就回滚
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			// 创建一个新的关系
			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetUser.ID
			contact.Type = 1
			// 数据库存入关系--> 因为是双向的，所以要添加两次
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败!"
			}
			contact1 := Contact{}
			contact1.OwnerId = targetUser.ID
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败!"
			}
			// 提交事务
			tx.Commit()

			return 0, "添加好友成功~"
		}
		return -1, "查无此用户!"
	}
	return -1, "好友用户名不能为空!"
}

// DeleteFriend
//
//	@Description: 删除好友关系
//	@param userId  当前用户id
//	@param targetName 目标用户用户名
//	@return int    0: 成功 1:失败
//	@return string 响应信息
func DeleteFriend(userId uint, targetName string) (int, string) {
	// 查找用户名
	if targetName == "" {
		return -1, "好友用户名不能为空!"
	}
	targetUser := FindUserByName(targetName) // 通过名字查找用户
	if targetUser.Salt == "" {
		return -1, "查无此用户！"
	}
	if targetUser.ID == userId {
		return -1, "不能添加自己！"
	}
	// 判断是否已经是好友关系
	contact0 := Contact{}
	utils.DB.Where("owner_id =?  and target_id =? and type=1", userId, targetUser.ID).Find(&contact0)
	if contact0.ID == 0 {
		return -1, "他/她并不是你的好友！"
	}
	contact1 := Contact{}
	utils.DB.Where("owner_id =?  and target_id =? and type=1", targetUser.ID, userId).Find(&contact1)
	if contact1.ID == 0 {
		return -1, "你并不是他/她的好友！"
	}
	// 开启一个事务
	tx := utils.DB.Begin()
	defer func() { //事务一旦开始，不论什么异常最终都会 Rollback
		// 如果出现异常就回滚
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := utils.DB.Delete(&contact0).Error; err != nil {
		tx.Rollback()
		return -1, "将它从你列表删除失败!"
	}
	if err := utils.DB.Delete(&contact1).Error; err != nil {
		tx.Rollback()
		return -1, "将你从它列表删除失败!"
	}
	// 提交事务
	tx.Commit()

	return 0, "删除好友成功~"
}

// SearchUserByGroupId
//
//	@Description: 查找当前群中的所有用户ID
//	@param communityId  群id
//	@return []uint 当前群中所有用户的ID
func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0) // 所有用户id
	// 条件查找 群id 并且 类型为群
	utils.DB.Where("target_id = ? and type=2", communityId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint(v.OwnerId))
	}
	return objIds
}
