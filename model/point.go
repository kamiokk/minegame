package model

import (
	"time"
	"github.com/kamiokk/minegame/helper/mysql"
)

// Point model of user's point
type Point struct {
	ID uint
	UserID uint `gorm:"column:user_id"`
	Point float32
	CreatedAt *time.Time
	UpdatedAt *time.Time
	IsDeleted uint
}

// TableName return user table name
func (Point) TableName() string {
	return "mine_user_point"
}

// GetByUserID fetch point by ID
func (p *Point) GetByUserID(id uint) {
	mysql.DBInstance().Where("user_id = ?", id).First(p)
}