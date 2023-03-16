package models

import (
	"gorm.io/gorm"
)

type MessageType int32

const (
	One2OneChat MessageType = 1 // 私聊
	GroupChat   MessageType = 2 // 群聊
	HeartBeat   MessageType = 3 //心跳检测
)

type InfoType int32

const (
	TextMsg    InfoType = 1 // 文本消息
	EmojiMsg   InfoType = 2 // 表情
	VoiceMsg   InfoType = 3 // 语音
	PictureMsg InfoType = 4 // 图片消息
)

// 消息
type Message struct {
	gorm.Model
	UserId     int64  //发送者
	TargetId   int64  //接受者
	Type       int    //发送类型  1私聊  2群聊  3心跳
	Media      int    //消息类型  1文字 2表情包 3语音 4图片
	Content    string //消息内容
	CreateTime uint64 //创建时间
	ReadTime   uint64 //读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}
