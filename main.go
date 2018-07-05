package main

import (
    "github.com/kamiokk/minegame/helper/mysql"
    redisHelper "github.com/kamiokk/minegame/helper/redis"
    "github.com/kamiokk/minegame/controller"
    "github.com/kamiokk/minegame/background"
)

func main()  {
    mysql.InitHelper()
    defer mysql.EndHelper()

    redisHelper.InitHelper(300)
    defer redisHelper.EndHelper()

    router := controller.Routers()

    background.InitBackground()

    router.Run(":80")
}