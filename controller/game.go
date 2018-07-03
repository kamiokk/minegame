package controller

import (
    "fmt"
    "time"
    "math/rand"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/gomodule/redigo/redis"
    redisHelper "github.com/kamiokk/minegame/helper/redis"
    "github.com/kamiokk/minegame/model"
    "github.com/kamiokk/minegame/helper/logHelper"
)

const (
    userPointLockPrefix = "lock:point:"
    gainRedPackLockPrefix = "lock:redpack:"

    minRedPackGainPoint uint = 25
    platformPercentage uint = 1
)

// GiveRedPack model
type GiveRedPack struct {
    Room uint `json:"room" binding:"required,min=1,max=6"`
    Mine uint `json:"mine" binding:"required,min=0,max=9"`
}

// RedPack struct
type RedPack struct {
    UserID uint
    Point uint
    Num uint
    Mine uint
    LossPay uint
    RemainPoint uint
    RemainNum uint
}

func poll(c *gin.Context) {
    c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"msg":MsgSucceed})
}

func giveOut(c *gin.Context)  {
    var json GiveRedPack
    if err := c.ShouldBindWith(&json,binding.JSON); err != nil {
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgBindJSONErr})
        return
    }
    giveOutPoint,giveOutNum,lossRatio := getRoomConfig(json.Room)
    if giveOutPoint == 0 || giveOutNum == 0 {
        logHelper.Warn(c,"GetRoomConfigFailed ID:%d",json.Room)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    session := startSession(c)
    //check user point
    userID,_ := session.GetUInt("userID")
    var p model.Point
    (&p).GetByUserID(userID)
    if p.Point * 100 < float64(giveOutPoint) {
        c.JSON(http.StatusOK,gin.H{"code":CodePointNotEngouth,"msg":MsgPointNotEngouth})
        return
    }
    rc := redisHelper.GetConn(c)
    lockKey := fmt.Sprintf(userPointLockPrefix + "%d",userID)
    lockID := redisHelper.RandLockId()
    lockOk := redisHelper.GetLockByTimeout(rc,time.Second * 5,lockKey,lockID,10)
    if !lockOk {
		logHelper.Warn(c,"GetLockFailed userID:%d lockKey:%s lockID:%v",userID,lockKey,lockID)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    defer redisHelper.ReleaseLock(rc,lockKey,lockID)
    //double check
    (&p).GetByUserID(userID)
    if p.Point * 100 < float64(giveOutPoint) {
        c.JSON(http.StatusOK,gin.H{"code":CodePointNotEngouth,"msg":MsgPointNotEngouth})
        return
    }
    rpModel := &model.RedPack {
        UserID: userID,
        Value: float64(giveOutPoint) / 100,
        MineNumber: json.Mine,
        Slice: giveOutNum,
    }
    rpModel.Create()
    if rpModel.ID <= 0 {
		logHelper.Warn(c,"CreateRedPackFailed UID:%d redpack:%v",userID,rpModel)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    if ok := (&p).ModifyPoint("-",float64(giveOutPoint) / 100);!ok {
		rpModel.IsDeleted = 1
		rpModel.Update()
        c.JSON(http.StatusOK,gin.H{"code":CodePointNotEngouth,"msg":MsgPointNotEngouth})
        return
    }
    pointAfterPer := giveOutPoint * (100 - platformPercentage) / 100
    rp := RedPack {
        UserID: userID,
        Point: pointAfterPer,
        Num: giveOutNum,
        Mine: json.Mine,
        LossPay: pointAfterPer * lossRatio / 10,
        RemainPoint: pointAfterPer,
        RemainNum: giveOutNum,
    }
    rpCacheKey := fmt.Sprintf("redpack:%d",rpModel.ID)
    if err := redisHelper.SetStructExp(rc,rpCacheKey,&rp,86400 * 2); err != nil {
		logHelper.Error(c,"RedisSetError val:%v",rp)
		(&p).ModifyPoint("+",float64(giveOutPoint) / 100)
		c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
		return
	}
    c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"id": rpModel.ID})
}

func gain(c *gin.Context) {
    var param struct{ID uint}
    if err := c.ShouldBindWith(&param,binding.JSON); err != nil {
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":err})
        return
    }
    rc := redisHelper.GetConn(c)
    // get redpack info
    var redpack RedPack
    rpCacheKey := fmt.Sprintf("redpack:%d",param.ID)
    if err := redisHelper.FetchStruct(rc,rpCacheKey,&redpack);err != nil {
        logHelper.Warn(c,"FetchRedPackFailed ID:%d error:%v",param.ID,err)
        c.JSON(http.StatusOK,gin.H{"code":CodeRedPackRunOut,"msg":MsgRedPackRunOut})
        return
    }
    if redpack.RemainNum <= 0 {
        logHelper.Warn(c,"RedPackRunOut ID:%d redpack:%v",param.ID,redpack)
        c.JSON(http.StatusOK,gin.H{"code":CodeRedPackRunOut,"msg":MsgRedPackRunOut})
        return
    }

    //lock user's point and check if user can afford loss
    session := startSession(c)
    userID,_ := session.GetUInt("userID")
    userLockKey := fmt.Sprintf(userPointLockPrefix + "%d",userID)
    userLockID := redisHelper.RandLockId()
    userLockOk := redisHelper.GetLockByTimeout(rc,time.Second * 5,userLockKey,userLockID,15)
    if !userLockOk {
        logHelper.Warn(c,"GetLockFailed userID:%d lockKey:%s lockID:%v",userID,userLockKey,userLockID)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    defer redisHelper.ReleaseLock(rc,userLockKey,userLockID)
    var p model.Point
    (&p).GetByUserID(userID)
    if p.Point * 100 < float64(redpack.LossPay) {
        c.JSON(http.StatusOK,gin.H{"code":CodeUnAffordRedPack,"msg":MsgUnAffordRedPack})
        return
    }

    // check if the user already get this redpack
    checkUserGainKey := fmt.Sprintf("gain:%d:%d",userID,param.ID)
    if check,err := redis.Int(rc.Do("incr",checkUserGainKey));err != nil {
        logHelper.Warn(c,"IncrFailed userID:%d redpackID:%d error:%v",userID,param.ID,err)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    } else {
        if check == 1 {
            rc.Do("expire",checkUserGainKey,86400 * 2)
        } else {
            c.JSON(http.StatusOK,gin.H{"code":CodeGained,"msg":MsgGained})
            return
        }
    }

    // lock this redpack
    lockKey := fmt.Sprintf(gainRedPackLockPrefix + "%d",param.ID)
    lockID := redisHelper.RandLockId()
    lockOk := redisHelper.GetLockByTimeout(rc,time.Second * 5,lockKey,lockID,10)
    if !lockOk {
        logHelper.Warn(c,"GetLockFailed userID:%d lockKey:%s lockID:%v",userID,lockKey,lockID)
        c.JSON(http.StatusOK,gin.H{"code":CodeRedPackRunOut,"msg":MsgRedPackRunOut})
        return
    }
    defer redisHelper.ReleaseLock(rc,lockKey,lockID)

    //double check 
    if err := redisHelper.FetchStruct(rc,rpCacheKey,&redpack);err != nil {
		logHelper.Warn(c,"FetchRedPackFailed ID:%d error:%v",param.ID,err)
		rc.Do("decr",checkUserGainKey)
        c.JSON(http.StatusOK,gin.H{"code":CodeRedPackRunOut,"msg":MsgRedPackRunOut})
        return
    }
    if redpack.RemainNum <= 0 || redpack.RemainPoint <= 0 {
		logHelper.Warn(c,"RedPackRunOut ID:%d redpack:%v",param.ID,redpack)
		rc.Do("decr",checkUserGainKey)
        c.JSON(http.StatusOK,gin.H{"code":CodeRedPackRunOut,"msg":MsgRedPackRunOut})
        return
    }
    redpack.RemainNum--
    var gainPoint uint
    if redpack.RemainNum == 0 {
        gainPoint = redpack.RemainPoint
        redpack.RemainPoint = 0
    } else {
        max := (redpack.RemainPoint / redpack.RemainNum) * 2
        gainPoint = uint(rand.Intn(int(max)))
        if gainPoint < minRedPackGainPoint {
            gainPoint = minRedPackGainPoint
        }
        if redpack.RemainPoint < (redpack.RemainNum * minRedPackGainPoint + gainPoint) {
            gainPoint = redpack.RemainPoint - (redpack.RemainNum * minRedPackGainPoint)
        }
        redpack.RemainPoint = redpack.RemainPoint - gainPoint
	}
	hitMine := checkHitMine(redpack.Mine,gainPoint)
    if hitMine {
        if err := model.TransferPoint(userID,redpack.UserID,float64(redpack.LossPay) / 100);err != nil {
			logHelper.Warn(c,"TransferPointError err:%v",err)
			rc.Do("decr",checkUserGainKey)
            c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
            return
        }
    } else {
		if ok := (&p).ModifyPoint("+",float64(gainPoint) / 100);!ok {
			logHelper.Warn(c,"GainPointError userID:%d gainPoint:%d redpack:%v",userID,gainPoint,redpack)
			rc.Do("decr",checkUserGainKey)
			c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
			return
		}
	}
	if err := redisHelper.SetStruct(rc,rpCacheKey,&redpack);err != nil {
		logHelper.Warn(c,"RedisSetError userID:%d gainPoint:%d redpack:%v error:%v",userID,gainPoint,redpack,err)
		if hitMine {
			model.TransferPoint(redpack.UserID,userID,float64(redpack.LossPay) / 100)
		} else {
			(&p).ModifyPoint("-",float64(gainPoint) / 100)
		}
		rc.Do("decr",checkUserGainKey)
		c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"gain": gainPoint,"hit": hitMine,"loss": redpack.LossPay})
}

func checkHitMine(mineNumber,gainPoint uint) bool {
	return gainPoint % 10 == mineNumber
}