package controller

import (
	"math"
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
    userPointLockKeyFmt = "lock:point:%d"
    gainRedPackLockKeyFmt = "lock:redpack:%d"
    pollLockKeyFmt = "lock:poll:%d"
    packSetKeyFmt = "set:pack:%d"
    checkGainKeyFmt = "gain:%d:%d"
    redpackKeyFmt = "redpack:%d"

    minRedPackGainPoint uint = 25
    platformPercentage uint = 1
    pollTimeout time.Duration = 60 * time.Second
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
    var param struct {
        Timestamp int64 `json:"t"`
        Room uint `json:"room" binding:"required,min=1,max=6"`
    }
    if err := c.ShouldBindWith(&param,binding.JSON); err != nil {
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgBindJSONErr})
        return
    }
    if param.Timestamp > time.Now().Unix() {
        c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"redpacks":nil,"t":param.Timestamp,"roomID":param.Room})
        return
    }

    session := startSession(c)
    userID,_ := session.GetUInt("userID")
    rc := redisHelper.GetConn(c)
    lockKey := fmt.Sprintf(pollLockKeyFmt,userID)
    lockID := redisHelper.RandLockId()
    lockOk := redisHelper.GetLockByTimeout(rc,time.Second * 5,lockKey,lockID,uint(pollTimeout + 10))
    if !lockOk {
		logHelper.Warn(c,"GetLockFailed userID:%d lockKey:%s lockID:%v",userID,lockKey,lockID)
        c.JSON(http.StatusOK,gin.H{"code":CodeEnterDupRoom,"msg":MsgEnterDupRoom})
        return
    }
    defer redisHelper.ReleaseLock(rc,lockKey,lockID)

    startAt := time.Now()
    if param.Timestamp > 0 && math.Abs(float64(startAt.Unix() - param.Timestamp)) < 10 {
        startAt = time.Unix(param.Timestamp,0)
    }
    rangeFrom := startAt.Unix()
    var rangeTo int64
    for time.Since(startAt) < pollTimeout {
        time.Sleep(time.Second * 1)
        rangeTo = time.Now().Unix()
        logHelper.Debug(c,"Polling: %d %d",rangeFrom,rangeTo)
        if packIds,err := redis.Int64s(rc.Do("zrangebyscore",fmt.Sprintf(packSetKeyFmt,param.Room),rangeFrom,rangeTo));err == nil && len(packIds) > 0 {
            logHelper.Debug(c,"Polling,packIds:%v",packIds)
            checkGainKeys := make([]interface{}, len(packIds))
            for i,packId := range packIds {
                checkGainKeys[i] = fmt.Sprintf(checkGainKeyFmt,userID,packId)
            }
            output := make(map[int64]RedPack, 0)
            if checkGain,err := redis.Int64s(rc.Do("mget",checkGainKeys...));err == nil {
                for i,gained := range checkGain {
                    if gained == 0 {
                        var rp RedPack
                        if redisHelper.FetchStruct(rc,fmt.Sprintf(redpackKeyFmt,packIds[i]),&rp) == nil {
                            output[packIds[i]] = rp
                        }
                    }
                }
            }
            logHelper.Debug(c,"Polling,output:%v",output)
            if len(output) > 0 {
                c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"redpacks":output,"t":(rangeTo+1),"roomID":param.Room})
                return
            }
        }
        rangeFrom = time.Now().Unix()
    }
    c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"redpacks":nil,"t":(rangeTo+1),"roomID":param.Room})
    return
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
    lockKey := fmt.Sprintf(userPointLockKeyFmt,userID)
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
    rpCacheKey := fmt.Sprintf(redpackKeyFmt,rpModel.ID)
    if err := redisHelper.SetStructExp(rc,rpCacheKey,&rp,86400 * 2); err != nil {
		logHelper.Error(c,"RedisSetError val:%v",rp)
		(&p).ModifyPoint("+",float64(giveOutPoint) / 100)
		c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
		return
    }
    if _,err := rc.Do("zadd",fmt.Sprintf(packSetKeyFmt,json.Room),time.Now().Unix(),rpModel.ID);err != nil {
        logHelper.Error(c,"RedisZaddError! score:%d value:%d",time.Now().Unix(),rpModel.ID)
        rc.Do("del",rpCacheKey)
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
    rpCacheKey := fmt.Sprintf(redpackKeyFmt,param.ID)
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
    userLockKey := fmt.Sprintf(userPointLockKeyFmt,userID)
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
    checkUserGainKey := fmt.Sprintf(checkGainKeyFmt,userID,param.ID)
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
    lockKey := fmt.Sprintf(gainRedPackLockKeyFmt,param.ID)
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