package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/kamiokk/minegame/model"
)

// Login post model
type Login struct {
	Account string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register post model
type Register struct {
	Account string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	PwdConf string `json:"pwd_conf" binding:"required,eqfield=Password"`
}

func login(c *gin.Context) {
	var json Login
	if err := c.ShouldBindWith(&json,binding.JSON); err == nil {
		var user model.User
		(&user).GetByAccount(json.Account)
		if user.ID == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 0,"msg": "user not found"})
		} else {
			if model.EncryptUserPassword(json.Password,user.PwdSalt) == user.Password {
				c.JSON(http.StatusOK, gin.H{"code": 1,"msg": "you are " + user.Account})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 0,"msg": "password wrong"})
			}
		}
	} else {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 0,"msg": "error"})
	}
}

func register(c *gin.Context) {
	var json Register
	if err := c.ShouldBindWith(&json,binding.JSON); err == nil {
		var user model.User
		(&user).GetByAccount(json.Account)
		if user.ID == 0 {
			user.Account = json.Account
			user.LastLoginIP = c.ClientIP()
			(&user).Create(json.Password)
			if user.ID > 0 {
				c.JSON(http.StatusOK, gin.H{"code": 1,"msg": "succeed"})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": 0,"msg": "create failed"})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 0,"msg": "user exists"})
		}
	} else {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 0})
	}
}