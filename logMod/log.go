package logMod

import (
	"SuperH/config"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

var CurrentLogDir *os.File

func Init() {

	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,                      //日志颜色
		FullTimestamp:          true,                      //日志事件
		TimestampFormat:        "2006-01-02 15:04:05 MST", //时间格式化
		DisableLevelTruncation: false,                     //日志等级简写
		DisableSorting:         true,                      //传入字段类型一致
	})

	//viper, _ := config.Instance()

	// 初始化日志配置
	path := config.Viper.GetString("LOG.path")
	//pathLink := config.Viper.GetString("LOG.pathlink")
	writer, err := rotatelogs.New(
		path+"%Y%m%d"+".log",
		rotatelogs.WithRotationCount(config.Viper.GetUint("LOG.maxday")),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		return
	}
	log.SetOutput(writer)
	log.AddHook(NewContextHook(log.InfoLevel))
	if config.ENV_IS_PRODUCTION {
		log.AddHook(NewContextHook(log.ErrorLevel))
	}
	log.Infof("log init success")
}

func GetTime() string {
	return time.Now().Format("2006-01-02")
}

func CreateDir(string2 string) error {
	err := os.MkdirAll(string2, 0755)
	if err != nil {
		return err
	}
	return nil
}

func DeleteOldLogFile() {
	dir := "./red-rain/logs"
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "du: %v\n", err)
		return
	}
	if len(entries) > 5 {
		for i := 0; i < len(entries)-5; i++ {
			err := os.Remove(dir + "/" + entries[i].Name())
			if err != nil {
				log.Errorf("log delete error %v", err)
			}
		}
	}
	return
}
