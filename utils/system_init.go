package utils

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func InitConfig() {
	viper.SetConfigName("app")    // 配置文件名称
	viper.AddConfigPath("config") // 查找配置文件所在的路径
	// 查找并读取文件
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	color.Green("Init Config Successfully!")
}

var (
	DB  *gorm.DB
	RDB *redis.Client
)

func InitMySQL() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢sql查询阈值
			LogLevel:      logger.Info, //级别
			Colorful:      true,
		},
	)

	var err error
	DB, err = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		color.Red("Connect Mysql Error!")
		panic(err)
	}
	color.Green("Init Mysql Successfully!")
}
func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConn"),
	})
	pong, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		color.Red("Init Redis Error : ", err)
		panic(err)
	} else {
		color.Green("Init Redis Successfully : %s", pong)
	}
}

const (
	PublishKey = "websocket"
)

// Publish 发布消息到Redis
func Publish(ctx context.Context, channel string, msg string) error {
	var err error
	fmt.Println("Publish 。。。。", msg)
	err = RDB.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Subscribe 订阅Redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := RDB.Subscribe(ctx, channel)
	fmt.Println("Subscribe1 ... ", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Subscribe2 ... ", msg.Payload)
	return msg.Payload, err
}
