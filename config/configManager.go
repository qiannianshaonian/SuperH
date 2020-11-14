/*
* 	config包利用viper读取配置文件
 */
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	ENV_IS_LOCAL      bool   = false
	ENV_IS_TEST       bool   = false
	ENV_IS_PRODUCTION bool   = false
	ENV               string = ""
	APP_PATH          string = ""
	CONFIG_PATH       string = ""
)

var Viper *viper.Viper

// 初始化配置对象，全局唯一
func Init() error {
	InitEnv()
	if Viper != nil {
		return nil
	}

	Viper = viper.New()

	Viper.SetConfigName(ENV)

	Viper.SetConfigType("yaml")

	Viper.AddConfigPath(CONFIG_PATH)

	if err := Viper.ReadInConfig(); err != nil {
		fmt.Println("init config error ", err)
		os.Exit(1)
		return err
	}

	return nil
}

// 获取viper实例
func Instance() (*viper.Viper, error) {
	if Viper == nil {
		err := Init()
		if err != nil {
			return nil, err
		}
	}

	return Viper, nil
}

func InitEnv() error {
	ENV = os.Getenv("AMONG_US_ENV")
	switch ENV {
	case "local":
		ENV_IS_LOCAL = true
	case "test":
		ENV_IS_TEST = true
	case "production":
		ENV_IS_PRODUCTION = true
	default:
		ENV = "local"
	}

	APP_PATH = "./"

	CONFIG_PATH = APP_PATH + "config/"

	return nil
}
