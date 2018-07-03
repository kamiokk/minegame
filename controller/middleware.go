package controller

import (
    "net/http"
    "github.com/gin-gonic/gin"
    redisHelper "github.com/kamiokk/minegame/helper/redis"
)

func checkLogin() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := startSession(c)
        if !isLogined(session) {
            c.JSON(http.StatusOK,gin.H{"code":CodeNeedLogin,"msg":MsgNeedLogin})
            c.Abort()
        }
        c.Next()
    }
}

func handlerEnd() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        redisHelper.CloseConn(c)
    }
}