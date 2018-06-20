package model

import (
	"time"
	"github.com/kamiokk/minegame/helper/mysql"
)

type User struct {
	ID uint
	Account string
	Password string
	PwdSalt string
	NickName string
	Phone string
	Email string
	Status uint
	LastLoginTime time.Time
	LastLoginIp string
	CreateeAt time.Time
	UpdatedAt time.Time
	IsDeleted uint
}

func (User) TableName() string {
	return "mine_user"
}

func (u *User) GetByID(id uint) {
	mysql.DBInstance().First(u,id)
}

func (u *User) GetByAccount(account string) {
	mysql.DBInstance().Where("account = ?", account).First(u)
}