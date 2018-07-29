package model

import (
	"math"
    "time"
    "github.com/kamiokk/minegame/helper/mysql"
)

// BalanceLog model of banlance log
type BalanceLog struct {
	ID uint
	EventID int
    UserID uint `gorm:"column:user_id"`
    RedpackID uint `gorm:"column:redpack_id"`
    Value float64
    CreatedAt *time.Time
    UpdatedAt *time.Time
    IsDeleted uint
}

// TableName return user table name
func (BalanceLog) TableName() string {
    return "mine_balance_log"
}

// Create add log
func (r *BalanceLog) Create(timestamp int64) {
	t := time.Unix(timestamp,0)
    r.CreatedAt = &t
    if mysql.DBInstance().NewRecord(r) {
        mysql.DBInstance().Create(r)
    }
}

// GetLogByUserID return user's logs
func GetLogByUserID(userID,offset,limit uint) (uint,[]BalanceLog) {
    var total uint
    var logs []BalanceLog
    mysql.DBInstance().Where("user_id=?",userID).Table("mine_balance_log").Count(&total)
    if offset < total {
        mysql.DBInstance().Where("user_id=?",userID).Offset(offset).Limit(limit).Order("id desc").Find(&logs)
    }
    return total,logs
}

type UserStat struct {
    Income float64
    GainCount uint
    GiveCount uint
    GainPoint float64
    GivePoint float64
    LoseTime uint
    WinTime uint
}

func StatByUserID(userID uint) UserStat {
    var ret UserStat
    var logs []BalanceLog
    mysql.DBInstance().Where("user_id=? and event_id not in (1,5)",userID).Find(&logs)
    if len(logs) > 0 {
        for _,val := range logs {
            ret.Income = floatCompute(ret.Income,val.Value,"+")
            switch val.EventID {
            case 2:
                ret.GiveCount++
                ret.GivePoint = floatCompute(ret.GivePoint,val.Value,"-") // 发红包Value为负数
            case 3:
                ret.GainCount++
                ret.GainPoint = floatCompute(ret.GainPoint,val.Value,"+")
            case 4:
                if val.Value > 0 {
                    ret.WinTime++
                } else {
                    ret.LoseTime++
                }
            }
        }
    }
    return ret
}

func floatCompute(a,b float64,op string) float64 {
    if op == "+" {
        return (math.Trunc(a * 100) + math.Trunc(b * 100)) / 100
    } else {
        return (math.Trunc(a * 100) - math.Trunc(b * 100)) / 100
    }
}