package controller

import (
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	router := gin.Default()
	groupUser := router.Group("/user")
	{
		groupUser.POST("/login",login)
		groupUser.GET("/register",register)
	}
	return router
}