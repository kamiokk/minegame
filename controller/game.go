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
    balanceEventCharge = 1
    balanceEventGiveOut = 2
    balanceEventGain = 3
    balanceEventHitMine = 4

    userPointLockKeyFmt = "lock:point:%d"
    gainRedPackLockKeyFmt = "lock:redpack:%d"
    pollLockKeyFmt = "lock:poll:%d"
    packSetKeyFmt = "set:pack:%d"
    endPackSetKeyFmt = "set:end:%d"
    gainSetKeyFmt = "set:gain:%d"
    checkGainKeyFmt = "gain:%d:%d"
    redpackKeyFmt = "redpack:%d"
    balanceLogQueue = "q:balance"

    minRedPackGainPoint uint = 25
    platformPercentage uint = 1
    agentPercentage uint = 30
    pollTimeout time.Duration = 60 * time.Second
    pollTimeoutS int = 60
    pollTimeDuration time.Duration = time.Second
)

type balance struct {
    UserID uint
    EventID int
    Value float64
    Timestamp int64
    RedpackID uint
}

// GiveRedPack model
type GiveRedPack struct {
    Room uint `json:"room" binding:"required,min=1,max=6"`
    Mine uint `json:"mine" binding:"required,min=0,max=9"`
}

// RedPack struct
type RedPack struct {
    UserID uint
    Account string
    Room uint
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
        c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"redpacks":nil,"ended":nil,"t":param.Timestamp,"roomID":param.Room})
        return
    }

    session := startSession(c)
    userID,_ := session.GetUInt("userID")
    rc := redisHelper.GetConn(c)
    lockKey := fmt.Sprintf(pollLockKeyFmt,userID)
    lockID := redisHelper.RandLockId()
    lockOk := redisHelper.GetLockByTimeout(rc,time.Second * 5,lockKey,lockID,uint(pollTimeoutS + 10))
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
        select {
        case <-c.Request.Context().Done():
            //when client abort the request,cancel polling
            logHelper.Debug(c,"poll canceled UID:%d",userID)
            c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"msg":"poll canceled"})
            return
        case <-time.After(pollTimeDuration):
            rangeTo = time.Now().Unix()
            logHelper.Debug(c,"Polling: %d [%d ~ %d]",userID,rangeFrom,rangeTo)
            //new red pack
            newRedpack := make(map[int64]RedPack, 0)
            packIDs,err := redis.Int64s(rc.Do("zrangebyscore",fmt.Sprintf(packSetKeyFmt,param.Room),rangeFrom,rangeTo));
            if err == nil && len(packIDs) > 0 {
                checkGainKeys := make([]interface{}, len(packIDs))
                for i,packID := range packIDs {
                    checkGainKeys[i] = fmt.Sprintf(checkGainKeyFmt,userID,packID)
                }
                if checkGain,err := redis.Int64s(rc.Do("mget",checkGainKeys...));err == nil {
                    for i,gained := range checkGain {
                        if gained == 0 {
                            var rp RedPack
                            if redisHelper.FetchStruct(rc,fmt.Sprintf(redpackKeyFmt,packIDs[i]),&rp) == nil {
                                newRedpack[packIDs[i]] = rp
                            }
                        }
                    }
                }
                logHelper.Debug(c,"Polling: %d [newRedpack:%v]",userID,newRedpack)
            }
            //ended red pack
            endRedpack := make(map[int64][]string, 0)
            endPackIDs,err := redis.Int64s(rc.Do("zrangebyscore",fmt.Sprintf(endPackSetKeyFmt,param.Room),rangeFrom,rangeTo));
            if err == nil && len(endPackIDs) > 0 {
                rpKeys := make([]interface{}, len(endPackIDs))
                for i,packID := range endPackIDs {
                    rpKeys[i] = fmt.Sprintf(redpackKeyFmt,packID)
                }
                if serializedRps,err := redis.ByteSlices(rc.Do("mget",rpKeys...));err == nil {
                    for i,srp := range serializedRps {
                        var rp RedPack
                        if err := redisHelper.Unmarshal(srp,&rp);err == nil {
                            endRedpack[endPackIDs[i]] = make([]string,0,rp.Num + 1)
                            endRedpack[endPackIDs[i]] = append(endRedpack[endPackIDs[i]],fmt.Sprintf("%s,%d,%d,%d",rp.Account,rp.Point,rp.LossPay,rp.Num))
                            if gainUsers,err := redis.Strings(rc.Do("spop",fmt.Sprintf(gainSetKeyFmt,endPackIDs[i]),rp.Num));err == nil {
                                for _,val := range gainUsers {
                                    endRedpack[endPackIDs[i]] = append(endRedpack[endPackIDs[i]],val)
                                }
                            }
                        }
                    }
                }
                logHelper.Debug(c,"Polling: %d [endRedPack:%v]",userID,endRedpack)
            }
            if len(newRedpack) > 0 || len(endRedpack) > 0 {
                c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"redpacks":newRedpack,"ended":endRedpack,"t":(rangeTo + 1),"roomID":param.Room})
                return
            }
            rangeFrom = rangeTo + 1
        }
    }
    c.JSON(http.StatusOK,gin.H{"code":CodeSucceed,"redpacks":nil,"ended":nil,"t":(rangeTo+1),"roomID":param.Room})
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
    agentID,_ := session.GetUInt("agentID")
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
    platfromValue := float64(giveOutPoint * platformPercentage) / 10000
    agentValue := float64(0)
    if agentID > 0 {
        agentValue = platfromValue * float64(agentPercentage) / 100
        platfromValue = (math.Trunc(platfromValue * 100) - math.Trunc(agentValue * 100)) / 100
    }
    rpModel := &model.RedPack {
        UserID: userID,
        Value: float64(giveOutPoint) / 100,
        PlatformValue: platfromValue,
        AgentValue: agentValue,
        MineNumber: json.Mine,
        Slice: giveOutNum,
    }
    rpModel.Create()
    if rpModel.ID <= 0 {
        logHelper.Warn(c,"CreateRedPackFailed UID:%d redpack:%v",userID,rpModel)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    if ok := (&p).ModifyPoint("-",rpModel.Value);!ok {
        rpModel.IsDeleted = 1
        rpModel.Update()
        c.JSON(http.StatusOK,gin.H{"code":CodePointNotEngouth,"msg":MsgPointNotEngouth})
        return
    }
    account,_ := session.GetString("account")
    pointAfterPer := giveOutPoint * (100 - platformPercentage) / 100
    rp := RedPack {
        UserID: userID,
        Account: account,
        Point: pointAfterPer,
        Num: giveOutNum,
        Mine: json.Mine,
        LossPay: pointAfterPer * lossRatio / 10,
        RemainPoint: pointAfterPer,
        RemainNum: giveOutNum,
        Room: json.Room,
    }
    rpCacheKey := fmt.Sprintf(redpackKeyFmt,rpModel.ID)
    if err := redisHelper.SetStructExp(rc,rpCacheKey,&rp,86400 * 2); err != nil {
        logHelper.Error(c,"RedisSetError val:%v",rp)
        (&p).ModifyPoint("+",rpModel.Value)
        c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
        return
    }
    if _,err := rc.Do("zadd",fmt.Sprintf(packSetKeyFmt,json.Room),time.Now().Unix() + 1,rpModel.ID);err != nil {
        logHelper.Error(c,"RedisZaddError! score:%d value:%d",time.Now().Unix() + 1,rpModel.ID)
        rc.Do("del",rpCacheKey)
        (&p).ModifyPoint("+",rpModel.Value)
        c.JSON(http.StatusOK, gin.H{"code": CodeFailed,"msg": MsgError})
        return
    }
    modifyBalanceLog(rc,userID,rpModel.ID,balanceEventGiveOut,rpModel.Value * -1)
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
    if checkGain,err := redis.Int(rc.Do("incr",checkUserGainKey));err != nil || checkGain != 1 {
        if err != nil {
            logHelper.Warn(c,"IncrFailed userID:%d redpackID:%d error:%v",userID,param.ID,err)
            c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        } else {
            c.JSON(http.StatusOK,gin.H{"code":CodeGained,"msg":MsgGained})
        }
        return
    }
    rc.Do("expire",checkUserGainKey,86400 * 2)

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
    lossPayFloat := float64(redpack.LossPay) / 100
    gainPointFloat := float64(gainPoint) / 100
    if ok := (&p).ModifyPoint("+",gainPointFloat);!ok {
        logHelper.Warn(c,"GainPointError userID:%d gainPoint:%d redpack:%v",userID,gainPoint,redpack)
        rc.Do("decr",checkUserGainKey)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    if hitMine {
        if err := model.TransferPoint(userID,redpack.UserID,lossPayFloat);err != nil {
            logHelper.Warn(c,"TransferPointError err:%v",err)
            (&p).ModifyPoint("-",gainPointFloat)
            rc.Do("decr",checkUserGainKey)
            c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
            return
        }
    }
    if err := redisHelper.SetStruct(rc,rpCacheKey,&redpack);err != nil {
        logHelper.Warn(c,"RedisSetError userID:%d gainPoint:%d redpack:%v error:%v",userID,gainPoint,redpack,err)
        if hitMine {
            model.TransferPoint(redpack.UserID,userID,lossPayFloat)
        }
        (&p).ModifyPoint("-",gainPointFloat)
        rc.Do("decr",checkUserGainKey)
        c.JSON(http.StatusOK,gin.H{"code":CodeFailed,"msg":MsgError})
        return
    }
    //finaly success
    account,_ := session.GetString("account")
    rc.Do("sadd",fmt.Sprintf(gainSetKeyFmt,param.ID),fmt.Sprintf("%s,%d,%v",account,gainPoint,hitMine))
    if redpack.RemainNum == 0 {
        rc.Do("zadd",fmt.Sprintf(endPackSetKeyFmt,redpack.Room),time.Now().Unix() + 1,param.ID)
    }
    if hitMine {
        transferLog(rc,userID,redpack.UserID,param.ID,lossPayFloat)
    }
    modifyBalanceLog(rc,userID,param.ID,balanceEventGain,gainPointFloat)
    c.JSON(http.StatusOK, gin.H{"code": CodeSucceed,"gain": gainPoint,"hit": hitMine,"loss": redpack.LossPay})
}

func checkHitMine(mineNumber,gainPoint uint) bool {
    return gainPoint % 10 == mineNumber
}

func transferLog(rc redis.Conn,srcID,dstID,rpID uint,value float64) {
    balanceLog := balance{
        UserID: srcID,
        EventID: balanceEventHitMine,
        Value: (value * -1),
        Timestamp: time.Now().Unix(),
        RedpackID: rpID,
    }
    redisHelper.LpushStruct(rc,balanceLogQueue,&balanceLog)
    balanceLog.UserID = dstID
    balanceLog.Value = value
    redisHelper.LpushStruct(rc,balanceLogQueue,&balanceLog)
}

func modifyBalanceLog(rc redis.Conn,userID,rpID uint,eventID int,value float64) {
    balanceLog := balance{
        UserID: userID,
        EventID: eventID,
        Value: value,
        Timestamp: time.Now().Unix(),
        RedpackID: rpID,
    }
    redisHelper.LpushStruct(rc,balanceLogQueue,&balanceLog)
}