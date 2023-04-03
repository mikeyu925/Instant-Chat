package models

import (
	"IM/utils"
	"context"
	"time"
)

/**
设置在线用户到redis缓存
**/
func SetUserOnlineInfo(key string, val []byte, timeTTL time.Duration) {
	ctx := context.Background()
	utils.RDB.Set(ctx, key, val, timeTTL)
}
