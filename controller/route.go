package controller

import (
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	router := gin.Default()
	//controllers 
	groupUser := router.Group("/user")
	{
		groupUser.POST("/login",login)
		groupUser.POST("/logout",logout)
		groupUser.POST("/register",register)
		groupUser.POST("/info",userInfo)

		groupUser.GET("/checkAccountAvailable",checkAccountAvailable)
	}

	//templates and static
	router.Static("/static", "./static")
	router.LoadHTMLGlob("./templates/*")
	router.GET("/", render)
	router.GET("/index.html", render)
	router.GET("/register.html", render)
	router.GET("/login.html", render)
	//router.GET("/test", render)
	return router
}