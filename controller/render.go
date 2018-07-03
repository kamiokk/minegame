package controller

import (
    "strings"
    "net/http"
    "github.com/gin-gonic/gin"
)

func render (c *gin.Context) {
    template := strings.Trim(c.Request.RequestURI,"/")
    if template == "" {
        template = "login.html"
    }
    c.HTML(http.StatusOK, template, gin.H{
        "title": "扫雷红包城",
    })
}