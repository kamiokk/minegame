package controller

import (
    "github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
    router := gin.Default()
    router.Use(handlerEnd())
    //controllers 
    groupUser := router.Group("/user")
    {
        groupUser.POST("/login",login)
        groupUser.POST("/logout",logout)
        groupUser.POST("/register",register)
        groupUser.POST("/info",userInfo)
        groupUser.GET("/checkAccountAvailable",checkAccountAvailable)
    }

    groupGame := router.Group("/game",checkLogin())
    {
        groupGame.POST("/poll",poll)
		groupGame.POST("/giveOut",giveOut)
		groupGame.POST("/gain",gain)
    }

    //templates and static
    router.Static("/static", "./static")
    router.LoadHTMLGlob("./templates/*")
    router.GET("/", render)
    router.GET("/index.html", render)
    router.GET("/register.html", render)
    router.GET("/login.html", render)
    router.GET("/test.html", render)
    return router
}