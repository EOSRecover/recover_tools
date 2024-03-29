package conf

import (
	"github.com/spf13/viper"
	"recover_tool/logger"
	"sync"
)

var once sync.Once
var appConfig APPConfig

type APPConfig struct {
	Debug          bool
	Node           string
	SendPrivateKey string
	SendAccount    string
	GMEndPoint     string
	
	HackerAccounts []string // hacker accounts
}

func APPConf() *APPConfig {
	
	once.Do(func() {
		
		if err := viper.Unmarshal(&appConfig); err != nil {
			
			logger.Instance().Error("read conf error -> ", err)
			return
		}
		
	})
	
	return &appConfig
}
