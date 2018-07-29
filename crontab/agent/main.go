package main

import (
	"math"
	"time"
	mysqlHelper "github.com/kamiokk/minegame/helper/mysql"
	"github.com/kamiokk/minegame/model"
)

func main() {
	mysqlHelper.InitHelper()
	defer mysqlHelper.EndHelper()

	var totalPerventValue float64
	var statTimeObj = time.Now()
	var statTime = statTimeObj.Format("2006-01-02 15:04:05")

	var lastAgentLog model.AgentLog
	mysqlHelper.DBInstance().Order("id desc").First(&lastAgentLog)
	var lastStatTime string
	if lastAgentLog.ID > 0 {
		lastStatTime = lastAgentLog.CreatedAt.Format("2006-01-02 15:04:05")
	} else {
		lastStatTime = "0000-00-00 00:00:00"
	}

	agentMap := make(map[uint][]uint)
    var users []model.User
	mysqlHelper.DBInstance().Where("status=1 and is_deleted=0").Find(&users)
	for _,user := range users {
		if user.AgentID > 0 {
			agentMap[user.AgentID] = append(agentMap[user.AgentID],user.ID)
		}
	}
	for agentID,userList := range agentMap {
		var percentValue float64
		var redpacks []model.RedPack
		mysqlHelper.DBInstance().Where("user_id in (?)",userList).Where("created_at > ?",lastStatTime).Where("created_at <= ?",statTime).Find(&redpacks)
		for _,redpack := range redpacks {
			percentValue = percentValue + math.Trunc(redpack.AgentValue * 100)
		}
		if percentValue > 0 {
			var userPoint model.Point
			userPoint.GetByUserID(agentID)
			userPoint.ModifyPoint("+",percentValue / 100)
			var log model.BalanceLog
			log.EventID = 6
			log.UserID = agentID
			log.Value = percentValue / 100
			log.Create(statTimeObj.Unix())
			totalPerventValue = totalPerventValue + percentValue
		}
	}
	var agentLog model.AgentLog
	agentLog.CreatedAt = &statTimeObj
	agentLog.Value = totalPerventValue / 100
	agentLog.Create()
}