package utils

import (
	"context"
	"fmt"
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
	viper.SetConfigName("app")
	viper.AddConfigPath("config")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println("config app:", viper.Get("app"))
	//fmt.Println("config mysql:", viper.Get("mysql"))
}

var (
	DB  *gorm.DB
	RDB *redis.Client
)

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
		fmt.Println("init redis error : ", err)
	} else {
		fmt.Println("init redis success : ", pong)
	}
}

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
		fmt.Println("连接数据库出错!")
		panic(err)
	}
	fmt.Println("mysql init finish!")
}
