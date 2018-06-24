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
	session := startSession(c)
	if isLogined(session) {
		c.JSON(http.StatusOK, gin.H{"code": CODE_LOGINED,"msg": MSG_LOGINED})
		return
	}
	var json Login
	if err := c.ShouldBindWith(&json,binding.JSON); err == nil {
		var user model.User
		(&user).GetByAccount(json.Account)
		if user.ID == 0 {
			c.JSON(http.StatusOK, gin.H{"code": CODE_LOGIN_ERR,"msg": MSG_LOGIN_ERR})
		} else {
			if model.EncryptUserPassword(json.Password,user.PwdSalt) == user.Password {
				err1 := session.Set("userID",user.ID)
				err2 := session.Set("account",user.Account)
				if err1 != nil || err2 != nil {
					c.JSON(http.StatusOK, gin.H{"code": CODE_FAILED,"msg": MSG_ERROR})
					return
				}
				c.JSON(http.StatusOK, gin.H{"code": CODE_SUCCEED,"msg": MSG_SUCCEED})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": CODE_LOGIN_ERR,"msg": MSG_LOGIN_ERR})
			}
		}
	} else {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"code": CODE_FAILED,"msg": MSG_ERROR})
	}
}

func register(c *gin.Context) {
	session := startSession(c)
	if isLogined(session) {
		c.JSON(http.StatusOK, gin.H{"code": CODE_LOGINED,"msg": MSG_LOGINED})
		return
	}
	var json Register
	if err := c.ShouldBindWith(&json,binding.JSON); err == nil {
		var user model.User
		(&user).GetByAccount(json.Account)
		if user.ID == 0 {
			user.Account = json.Account
			user.LastLoginIP = c.ClientIP()
			(&user).Create(json.Password)
			if user.ID > 0 {
				c.JSON(http.StatusOK, gin.H{"code": CODE_SUCCEED,"msg": MSG_SUCCEED})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": CODE_REGISTER_ERR,"msg": MSG_REGISTER_ERR})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"code": CODE_REGISTER_DUP_NAME,"msg": MSG_REGISTER_DUP_NAME})
		}
	} else {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"code": CODE_FAILED})
	}
}

func checkAccountAvailable(c *gin.Context) {
	account := c.Query("account")
	if account == "" {
		c.JSON(http.StatusOK, gin.H{"code": CODE_FAILED,"msg": MSG_ERROR})
		return
	}
	var user model.User
	(&user).GetByAccount(account)
	if user.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": CODE_SUCCEED,"msg": MSG_SUCCEED})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": CODE_FAILED,"msg": MSG_ERROR})
	}
}

func logout(c *gin.Context) {
	session := startSession(c)
	if isLogined(session) {
		session.Unset("userID")
		session.Unset("account")
		c.JSON(http.StatusOK, gin.H{"code": CODE_SUCCEED,"msg": MSG_SUCCEED})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": CODE_FAILED,"msg": MSG_ERROR})
}

func userInfo(c *gin.Context) {
	session := startSession(c)
	if isLogined(session) {
		userID,_ := session.GetUInt("userID")
		account,_ := session.GetString("account")
		var p model.Point
		(&p).GetByUserID(uint(userID))
		c.JSON(http.StatusOK, gin.H{"code": CODE_SUCCEED,"id": userID,"account": account,"point": p.Point})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": CODE_NEED_LOGIN,"msg": MSG_NEED_LOGIN})
}