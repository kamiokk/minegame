package controller

import (
    "sync"
    "github.com/kamiokk/gosession"
    memsess "github.com/kamiokk/gosession/mem"
	"github.com/gin-gonic/gin"
	"github.com/kamiokk/minegame/helper/logHelper"
)

const (
    CodeFailed = 0
    CodeSucceed = 1
    CodeNeedLogin = 10
    CodeLoginErr = 11
    CodeLogined = 12
    CodeRegisterDupName = 20
    CodeRegisterErr = 21
    CodePointNotEngouth = 30
    CodeRedPackRunOut = 40
    CodeUnAffordRedPack = 41
    CodeGained = 42
    CodeHitMine = 43
    CodeEnterDupRoom = 44
)

const (
    MsgSucceed = "succeed"
    MsgError = "error"
    MsgNeedLogin = "you need to login first"
    MsgLoginErr = "account or password wrong"
    MsgLogined = "already logined"
    MsgRegisterDupName = "account already exist"
    MsgRegisterErr = "register failed"
    MsgPointNotEngouth = "point not engouth"
    MsgBindJSONErr = "binding json error"
    MsgRedPackRunOut = "redpack run out"
    MsgUnAffordRedPack = "can not afford red pack"
    MsgGained = "already gained"
    MsgHitMine = "hit the mine"
    MsgEnterDupRoom = "already enter another room"
)

var roomConfigOnce sync.Once
var roomConfig map[uint][3]uint

func startSession(c *gin.Context) (*gosession.Session) {
    if s,ok := c.Get("__sessionPointer");ok {
        _,ok = s.(*gosession.Session)
        if ok {
			logHelper.Debug(c,"GetSessionFromContext:%v",s)
            return s.(*gosession.Session)
        }
    }

    smodel := &memsess.Model{}
    session,err := gosession.Start(c.Request,c.Writer,smodel)
    if err != nil {
        panic("Can not start a new session.")
    }
	c.Set("__sessionPointer",session)
	logHelper.Debug(c,"StartNewSession:%v",session)
    return session
}

func isLogined(s *gosession.Session) bool {
    uid,err := s.GetUInt("userID")
    if uid > 0 {
        return true
    }
    if err != nil {
		logHelper.DebugNoContext("CheckLoginErr:%v",err)
    }
    return false
}

func getRoomConfig(number uint) (uint,uint,uint) {
    roomConfigOnce.Do(initRoomConfig)
    if _,ok := roomConfig[number];!ok {
        return 0,0,0
    }
    return roomConfig[number][0],roomConfig[number][1],roomConfig[number][2]
}

func initRoomConfig() {
    roomConfig = make(map[uint][3]uint,6)
    roomConfig[1] = [3]uint{3000,7,15}
    roomConfig[2] = [3]uint{5000,7,15}
    roomConfig[3] = [3]uint{10000,7,15}
    roomConfig[4] = [3]uint{20000,7,15}
    roomConfig[5] = [3]uint{50000,7,15}
    roomConfig[6] = [3]uint{100000,7,15}
}