package config

import (
	"github.com/jinzhu/gorm"
	_ "gorm.io/driver/mysql"
)

var DB *gorm.DB

func init() {
	var err error
	DB, err = gorm.Open("mysql", "root:admin123@tcp(127.0.0.1:3306)/wallet_signer?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}
}
