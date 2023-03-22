package main

import (
	"IM/models"
	"IM/utils"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	// Migrate the schema
	//utils.DB.AutoMigrate(&models.UserBasic{})
	//utils.DB.AutoMigrate(&models.Message{})
	//utils.DB.AutoMigrate(&models.GroupBasic{})
	//utils.DB.AutoMigrate(&models.Contact{})
	utils.DB.AutoMigrate(&models.Community{})
}
