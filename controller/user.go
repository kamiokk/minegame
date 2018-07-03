package controller

import (
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
        c.JSON(http.StatusOK, gin.H{"code": CodeLogined,"msg": MsgLogined})
        return
    }
    var json Login
    if err := c.ShouldBindWith(&json,binding.JSON); err == nil {
        var user model.User
        (&user).GetByAccount(json.Account)
        if user.ID == 0 {
            c.JSON(http.StatusOK, gin.H{"code": CodeLoginErr,"msg": MsgLoginErr})
        } else {
            if model.EncryptUserPassword(json.Password,user.PwdSalt) == user.Password {
                err1 := session.Set("userID",user.ID)
                err2 := session.Set("account",user.Account)
                if err1 != nil || err2 != nil {
                    c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
                    return
                }
                c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"msg": MsgSucceed})
            } else {
                c.JSON(http.StatusOK, gin.H{"code": CodeLoginErr,"msg": MsgLoginErr})
            }
        }
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"code": CodeFailed,"msg": MsgError})
    }
}

func register(c *gin.Context) {
    session := startSession(c)
    if isLogined(session) {
        c.JSON(http.StatusOK, gin.H{"code": CodeLogined,"msg": MsgLogined})
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
                c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"msg": MsgSucceed})
            } else {
                c.JSON(http.StatusOK, gin.H{"code": CodeRegisterErr,"msg": MsgRegisterErr})
            }
        } else {
            c.JSON(http.StatusOK, gin.H{"code": CodeRegisterDupName,"msg": MsgRegisterDupName})
        }
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"code": CodeFailed})
    }
}

func checkAccountAvailable(c *gin.Context) {
    account := c.Query("account")
    if account == "" {
        c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
        return
    }
    var user model.User
    (&user).GetByAccount(account)
    if user.ID == 0 {
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"msg": MsgSucceed})
    } else {
        c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
    }
}

func logout(c *gin.Context) {
    session := startSession(c)
    if isLogined(session) {
        session.Unset("userID")
        session.Unset("account")
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"msg": MsgSucceed})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
}

func userInfo(c *gin.Context) {
    session := startSession(c)
    if isLogined(session) {
        userID,_ := session.GetUInt("userID")
        account,_ := session.GetString("account")
        var p model.Point
        (&p).GetByUserID(userID)
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"id": userID,"account": account,"point": p.Point})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": CodeNeedLogin,"msg": MsgNeedLogin})
}