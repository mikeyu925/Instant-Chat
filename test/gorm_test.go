package test

import (
	"IM/models"
	"IM/utils"
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGorm(t *testing.T) {
	utils.InitConfig()
	db, err := gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{})
	if err != nil {
		fmt.Println("连接数据库出错")
		panic(err)
	}
	// Migrate the schema
	db.AutoMigrate(&models.UserBasic{})

	// Create
	user := &models.UserBasic{}
	user.Name = "root"
	user.PassWord = "123456"
	db.Create(&user)

	// Read
	fmt.Println(db.First(user, 1)) // find product with integer primary key

	// Update - update product's price to 200
	//db.Model(&user).Update("PassWord", "123")
	// Update - update multiple fields
	//db.Model(&user).Updates(models.UserBasic{Price: 200, Code: "F42"}) // non-zero fields
	//db.Model(&user).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	db.Delete(&user, 1)
}