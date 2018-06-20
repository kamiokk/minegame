package mysql

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func init() {
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
}

func DBInstance() *gorm.DB {
	return db
}