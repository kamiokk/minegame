package controller

import (
    "strings"
    "net/http"
    "github.com/gin-gonic/gin"
)

var tpls []string

func init()  {
    tpls = []string{
    "index.html",
    "register.html",
    "login.html",
    "rooms.html",
    "room.html",
    "stat.html",
    "balanceDetail.html",
    "account.html",
    "agent.html",
    "test.html",
    }
}

func InitRender(router *gin.Engine) {
    router.GET("/", render)
    for _,tpl := range tpls {
        router.GET("/" + tpl, render)
    }
}

func render (c *gin.Context) {
    var template string
    uri := strings.Split(strings.Trim(c.Request.RequestURI,"/"),"?")
    if len(uri) > 0 {
        for _,tpl := range tpls {
            if uri[0] == tpl {
                template = tpl
                break;
            }
        }
    }
    if template == "" {
        template = "login.html"
    }
    c.HTML(http.StatusOK, template, gin.H{
        "title": "鸿燚娱乐城",
    })
}