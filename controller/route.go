package controller

import (
    "github.com/gin-gonic/gin"
)

// Routers return gin routers
func Routers() *gin.Engine {
    router := gin.Default()
    router.Use(handlerEnd())
    //controllers 
    groupUser := router.Group("/user")
    {
        groupUser.GET("/isLogin",isLogin)
        groupUser.POST("/login",login)
        groupUser.POST("/logout",logout)
        groupUser.POST("/register",register)
        groupUser.POST("/info",userInfo)
        groupUser.GET("/checkAccountAvailable",checkAccountAvailable)
        groupUser.GET("/balanceLog",balanceLog)
        groupUser.GET("/stat",stat)
        groupUser.POST("/agentCount",agentCount)
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
    InitRender(router)
    return router
}