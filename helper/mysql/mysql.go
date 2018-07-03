package mysql

import (
    "time"
    "os"

    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func InitHelper() {
    host := os.Getenv("MYSQL_PORT_3306_TCP_ADDR")
    databaseName := os.Getenv("MYSQL_ENV_MYSQL_DATABASE")
    user := os.Getenv("MYSQL_ENV_MYSQL_USER")
    password := os.Getenv("MYSQL_ENV_MYSQL_PASSWORD")
    dsn := user+":"+password+"@tcp("+host+")/"+databaseName+"?charset=utf8&parseTime=True&loc=Local"
    var err error
    db, err = gorm.Open("mysql", dsn)
    if err != nil {
        panic(err)
    }
    db.DB().SetMaxIdleConns(30)
    db.DB().SetMaxOpenConns(30)
    db.DB().SetConnMaxLifetime(time.Second * 60)
}

func EndHelper() {
    if db != nil {
        db.Close()
    }
}

func DBInstance() *gorm.DB {
    return db
}