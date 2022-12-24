package mysql

import (
	"fmt"

	"github.com/shinemost/grpc-up/models"
	"github.com/shinemost/grpc-up/settings"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() error {

	my := settings.Cfg.Mysqls

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", my.User, my.Password, my.Host, my.Port, my.Dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("gorm init error", err)
		return err
	}

	err = db.AutoMigrate(&models.Product{})
	if err != nil {
		fmt.Println("create database tables error", err)
		return err
	}

	return nil

}
