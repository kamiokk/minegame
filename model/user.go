package model

import (
    "os"
    "fmt"
    "crypto/sha1"
    "time"
    "math/rand"
    "github.com/kamiokk/minegame/helper/mysql"
)

// User model of user
type User struct {
    ID uint
    Account string
    Password string
    PwdSalt string
    NickName string
    Phone string
    Email string
    Status uint `gorm:"not null;default:1"`
    LastLoginTime *time.Time
    LastLoginIP string `gorm:"column:last_login_ip"`
    CreatedAt *time.Time
    UpdatedAt *time.Time
    IsDeleted uint
}

// EncryptUserPassword generate encrypted password string to store in DB
func EncryptUserPassword(password,salt string) string {
    s1 := sha1.New()
    s1.Write([]byte(password + os.Getenv("MINE_USER_PWD_SECRET") + salt))
    encrypt := fmt.Sprintf("%x",s1.Sum(nil))
    return encrypt
}

// TableName return user table name
func (User) TableName() string {
    return "mine_user"
}

// GetByID fetch user by ID
func (u *User) GetByID(id uint) {
    mysql.DBInstance().First(u,id)
}

// GetByAccount fetch user by account
func (u *User) GetByAccount(account string) {
    mysql.DBInstance().Where("account = ?", account).First(u)
}

// Create add new user
func (u *User) Create(rawPwd string) {
    u.PwdSalt = randomString(8)
    u.Password = EncryptUserPassword(rawPwd,u.PwdSalt)
    if mysql.DBInstance().NewRecord(u) {
        mysql.DBInstance().Create(u)
    }
}

func randomString(length int) string {
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    ret := make([]rune,length)
    for i := range ret {
        ret[i] = letters[rand.Intn(len(letters))]
    }
    return string(ret)
}