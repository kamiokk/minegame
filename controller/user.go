package controller

import (
	"strconv"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/kamiokk/minegame/model"
    //"github.com/kamiokk/minegame/helper/logHelper"
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
    Agent string `json:"agent"`
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
                err3 := session.Set("agentID",user.AgentID)
                if err1 != nil || err2 != nil || err3 != nil {
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
    var reqData Register
    if err := c.ShouldBindWith(&reqData,binding.JSON); err == nil {
        var user model.User
        (&user).GetByAccount(reqData.Account)
        if user.ID == 0 {
            user.Account = reqData.Account
            user.LastLoginIP = c.ClientIP()
            if reqData.Agent != "" {
                var agent model.User
                (&agent).GetByAccount(reqData.Agent)
                user.AgentID = agent.ID
            }
            (&user).Create(reqData.Password)
            if user.ID > 0 {
                var userPoint model.Point
                userPoint.UserID = user.ID
                userPoint.Point = 0
                userPoint.Create()
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

func balanceLog(c *gin.Context) {
    queryOffset := c.DefaultQuery("offset", "0")
    var offset uint64
    offset,err := strconv.ParseUint(queryOffset,10,0)
    if err != nil {
        offset = 0
    }
    var pageSize uint = 10
    session := startSession(c)
    if isLogined(session) {
        type bLog struct{
            Point float64 `json:"point"`
            Event string `json:"event"`
            CreateAt string `json:"time"`
        }
        var logs []bLog
        userID,_ := session.GetUInt("userID")
        total,balanceLog := model.GetLogByUserID(userID,uint(offset),pageSize)
        isLast := uint(offset) + pageSize >= total
        if total > 0 && len(balanceLog) > 0 {
            logs = make([]bLog, len(balanceLog), len(balanceLog))
            for i,v := range balanceLog {
                logs[i].Point = v.Value
                switch v.EventID {
                case 1:
                    logs[i].Event = "充值"
                case 2:
                    logs[i].Event = "发红包"
                case 3:
                    logs[i].Event = "抢红包"
                case 4:
                    logs[i].Event = "地雷"
                case 5:
                    logs[i].Event = "提现"
                case 6:
                    logs[i].Event = "代理提成"
                }
                logs[i].CreateAt = v.CreatedAt.Format("2006-01-02 15:04:05")
            }
        }
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"logs": logs,"last": isLast})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": CodeNeedLogin,"msg": MsgNeedLogin})
}

func stat(c *gin.Context) {
    session := startSession(c)
    if isLogined(session) {
        userID,_ := session.GetUInt("userID")
        stat := model.StatByUserID(userID)
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"stat": stat})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": CodeNeedLogin,"msg": MsgNeedLogin})
}

func isLogin(c *gin.Context)  {
    session := startSession(c)
    if isLogined(session) {
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"msg": MsgSucceed})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
}

func agentCount(c *gin.Context) {
    session := startSession(c)
    if isLogined(session) {
        userID,_ := session.GetUInt("userID")
        var user model.User
        c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"count": user.CountAgent(userID)})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": CodeNeedLogin,"msg": MsgNeedLogin})
}