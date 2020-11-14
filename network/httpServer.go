package network

import (
	"SuperH/config"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

func InitHttpServer() {
	viper, _ := config.Instance()
	port := viper.GetString("HTTP_PORT")
	//http.HandleFunc("/api/redrain/list", api.GetRedRainListApi)
	url := fmt.Sprintf("0.0.0.0:%s", port)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		logrus.Infof("http server start error %v", err)
	}
}
