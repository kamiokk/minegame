package model

import (
	"fmt"
    "time"
	"github.com/jinzhu/gorm"
	"github.com/kamiokk/minegame/helper/mysql"
)

// Point model of user's point
type Point struct {
    ID uint
    UserID uint `gorm:"column:user_id"`
    Point float64
    CreatedAt *time.Time
    UpdatedAt *time.Time
    IsDeleted uint
}

// TableName return user table name
func (Point) TableName() string {
    return "mine_user_point"
}

// Create add record
func (r *Point) Create() {
	now := time.Now()
    r.CreatedAt = &now
    if mysql.DBInstance().NewRecord(r) {
        mysql.DBInstance().Create(r)
    }
}

// GetByUserID fetch point by ID
func (p *Point) GetByUserID(id uint) {
    mysql.DBInstance().Where("user_id = ?", id).First(p)
}

// ModifyPoint modify user's point safely
func (p *Point) ModifyPoint(op string,value float64) bool {
	var affected int64
	if op == "+" {
		affected = mysql.DBInstance().Model(p).Update("point", gorm.Expr("point + ?", value)).RowsAffected
	}
	if op == "-" {
		affected = mysql.DBInstance().Model(p).Where("point >= ?", value).Update("point", gorm.Expr("point - ?", value)).RowsAffected
	}
	return affected > 0
}

// TransferPoint from src user to dst user
func TransferPoint(srcID,dstID uint,value float64) error {
	// Note the use of tx as the database handle once you are within a transaction
	tx := mysql.DBInstance().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	} ()

	if tx.Error != nil {
		return tx.Error
	}

	affected := tx.Table("mine_user_point").Where("user_id=?",srcID).Where("point >= ?", value).Update("point", gorm.Expr("point - ?", value)).RowsAffected
	if !(affected > 0) {
		tx.Rollback()
		return fmt.Errorf("DecreasePointFailed UID:%d Value:%f Error:%v",srcID,value,tx.Error)
	}

	affected2 := tx.Table("mine_user_point").Where("user_id=?",dstID).Update("point", gorm.Expr("point + ?", value)).RowsAffected
	if !(affected2 > 0) {
		tx.Rollback()
		return fmt.Errorf("IncreasePointFailed UID:%d Value:%f Error:%v",dstID,value,tx.Error)
	}

	err := tx.Commit().Error
	return err
}