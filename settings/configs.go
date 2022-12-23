package settings

import (
	"fmt"

	"github.com/shinemost/grpc-up/models"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitConfig() error {

	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.SetConfigName("configs")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("viper init error", err)
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", viper.GetString("mysql.user"), viper.GetString("mysql.password"), viper.GetString("mysql.host"), viper.GetString("mysql.port"), viper.GetString("mysql.dbname"))
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

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
