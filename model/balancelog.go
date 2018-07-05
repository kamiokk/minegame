package model

import (
    "time"
    "github.com/kamiokk/minegame/helper/mysql"
)

// BalanceLog model of banlance log
type BalanceLog struct {
	ID uint
	EventID int
    UserID uint `gorm:"column:user_id"`
    Value float64
    CreatedAt *time.Time
    UpdatedAt *time.Time
    IsDeleted uint
}

// TableName return user table name
func (BalanceLog) TableName() string {
    return "mine_banlance_log"
}

// Create add log
func (r *BalanceLog) Create(timestamp int64) {
	t := time.Unix(timestamp,0)
    r.CreatedAt = &t
    if mysql.DBInstance().NewRecord(r) {
        mysql.DBInstance().Create(r)
    }
}