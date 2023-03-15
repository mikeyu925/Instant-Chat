package main

import (
	"IM/router"
	"IM/utils"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	r := router.Router()
	r.Run() // listen and serve
}
