package models

import (
	"IM/utils"
	"context"
	"time"
)

// SetUserOnlineInfo
//
//	@Description: 设置在线用户到redis缓存
//	@param key ："online_"+Idstr
//	@param val ： 用户地址 []byte(node.Addr)
//	@param timeTTL  过期时间 默认4小时
func SetUserOnlineInfo(key string, val []byte, timeTTL time.Duration) {
	ctx := context.Background()
	utils.RDB.Set(ctx, key, val, timeTTL)
}
