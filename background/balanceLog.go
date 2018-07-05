package background

import (
	"time"
	redisHelper "github.com/kamiokk/minegame/helper/redis"
	"github.com/kamiokk/minegame/helper/logHelper"

	"github.com/kamiokk/minegame/model"
)

type balance struct {
	UserID uint
	EventID int
	Value float64
	Timestamp int64
}

// HandleBalanceLogQueue adding log to db
func HandleBalanceLogQueue() {
	rc := redisHelper.Pool().Get()
	sleepDuration := time.Millisecond * 10
	for {
		time.Sleep(sleepDuration)
		var b balance
		popSucceed,err := redisHelper.LpopStruct(rc,"q:balance",b)
		if !popSucceed {
			if sleepDuration < time.Second {
				sleepDuration += time.Millisecond * 10
			}
			continue
		} else {
			sleepDuration = time.Millisecond * 10
		}
		if err != nil {
			logHelper.DebugNoContext("HandleBalanceLogQueue:%v",err)
			continue;
		}
		log := model.BalanceLog{
			EventID: b.EventID,
			UserID: b.UserID,
			Value: b.Value,
		}
		log.Create(b.Timestamp)
	}
}