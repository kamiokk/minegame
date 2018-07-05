package redpack

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"math/rand"
	redisHelper "github.com/kamiokk/minegame/helper/redis"
)

type RedPack struct {
	ID int
	PointX100 int
	Remain int
	Mine int8
}

func (r *RedPack) new(c *gin.Context,userID int) bool {
	rc := redisHelper.GetConn(c)
	lockKey := fmt.Sprintf("lock:point:%d",userID)
	lockID := time.Now().UnixNano() * 10000 + rand.Int63n(10000)
	succeed := redisHelper.GetLockByTimeout(rc,time.Second * 5,lockKey,lockID,time.Second * 5)
	if succeed {
		fmt.Println("new red pack")
		redisHelper.ReleaseLock(rc,lockKey,lockID)
		return true
	}
	return false
}

func GetRedpack(redpackID int) {
	
}