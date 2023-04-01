package main

import (
	"IM/router"
	"IM/utils"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()
	r := router.Router()
	r.Run()
}
