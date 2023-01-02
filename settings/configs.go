package settings

import (
	"fmt"

	"github.com/spf13/viper"
)

type Configs struct {
	Web
	Mysqls
	Grpc
}

type Mysqls struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

type Web struct {
	Name    string
	Mod     string
	Port    string
	CrtFile string
	KeyFile string
	CaFile  string
}

type Grpc struct {
	Address string
}

var Cfg Configs

func InitConfigs() error {
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.SetConfigName("configs")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("viper init error", err)
		return err
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		fmt.Printf("unable to decode into struct, %v", err)
		return err
	}

	return nil

}
