package controller

import (
	"fmt"
	"log"
	"crypto/sha1"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/kamiokk/minegame/model"
)

type Login struct {
	Account string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func login(c *gin.Context) {
	var json Login
	if err := c.ShouldBindWith(&json,binding.JSON); err == nil {
		var user model.User
		(&user).GetByAccount(json.Account)
		if user.ID == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 0,"msg": "user not found"})
		} else {
			s1 := sha1.New()
			s1.Write([]byte(json.Password + "Secret@mine@2018" + user.PwdSalt))
			encrypt := fmt.Sprintf("%x",s1.Sum(nil))
			log.Println(user.Password)
			log.Println(encrypt)
			if encrypt == user.Password {
				c.JSON(http.StatusOK, gin.H{"code": 1,"msg": "you are " + user.Account})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 0,"msg": "password wrong"})
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0})
	}
}

func register(c *gin.Context) {
	c.String(http.StatusOK,"this is register interface.")
}