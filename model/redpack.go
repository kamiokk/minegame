package model

import (
    "time"
    "github.com/kamiokk/minegame/helper/mysql"
)

// RedPack model
type RedPack struct {
    ID uint
    UserID uint `gorm:"column:user_id"`
    Value float64
    MineNumber uint
    Slice uint
    CreatedAt *time.Time
    UpdatedAt *time.Time
    IsDeleted uint
}

// TableName return table name
func (RedPack) TableName() string {
    return "mine_red_pack"
}

// Create add record
func (r *RedPack) Create() {
    if mysql.DBInstance().NewRecord(r) {
        mysql.DBInstance().Create(r)
    }
}

// Update record
func (r *RedPack) Update() error {
	return mysql.DBInstance().Save(r).Error
}