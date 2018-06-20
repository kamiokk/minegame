package main

import (
	"github.com/kamiokk/minegame/helper/mysql"
	"github.com/kamiokk/minegame/controller"
)

func main()  {
	defer mysql.DBInstance().Close()
	router := controller.Routers()
	router.Run(":80")
}