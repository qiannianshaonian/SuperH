package myMysql

import (
	"SuperH/config"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync/atomic"
	"time"
)

var myDb *gorm.DB
var qPDb *gorm.DB // 气泡从库

var maxRedRainId int64

var ConfigMap = make(map[string]int64)

type Handlers func(*gorm.DB) error

func GetDb() *gorm.DB {
	return myDb
}

func GetNextRedRainId() int64 {
	return atomic.AddInt64(&maxRedRainId, 1)
}

func InitMaxRedRainId() {
	redRainId := GetRedRainIdMax() //
	atomic.StoreInt64(&maxRedRainId, redRainId)
}

type Model struct {
	Id        int64 `gorm:"primary_key" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func InitMysqlDb() {
	viper, _ := config.Instance()
	var err error
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // 慢 SQL 阈值
			LogLevel:      logger.Silent, // Log level
			Colorful:      false,         // 禁用彩色打印
		},
	)
	user := viper.GetString("DATABASE.user")
	pwd := viper.GetString("DATABASE.pwd")
	host := viper.GetString("DATABASE.host")
	port := viper.GetString("DATABASE.port")
	dbName := viper.GetString("DATABASE.name")
	maxconns := viper.GetInt("DATABASE.maxconns")
	idleconns := viper.GetInt("DATABASE.idleconns")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pwd, host, port, dbName)
	myDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		logrus.Errorf("init database err: [%v]", err)
		return
	}

	db, err := myDb.DB()
	if err != nil {
		logrus.Errorf("init database err: [%v]", err)
		return
	}
	db.SetMaxIdleConns(idleconns)
	db.SetMaxOpenConns(maxconns)
	if err := myDb.AutoMigrate(
		new(RedRains), new(PlayerRedRainsLogs),
	); err != nil {
		logrus.Infof("db auto sync err: %v", err)
	}
}

func Transaction(handlers []Handlers) (err error) {
	txSession := myDb.Begin()
	if err = txSession.Error; err != nil {
		log.Printf("db transaction begin error: %v", err)
		return
	}
	defer func() {
		if err != nil {
			log.Printf("transaction error: %v", err)
			txSession.Rollback()
		}
	}()
	for i := 0; i < len(handlers); i++ {
		if err = handlers[i](txSession); err != nil {
			return
		}
	}
	return txSession.Commit().Error
}

func UpMysql(values []interface{}, mysqlTab string) {
	switch mysqlTab {
	case "red_rain":
		dealRedRain(values)
	case "player_red_rain":
		dealPlayerRedRain(values)
	default:
	}
}

func IsNeedUpTable(tableName string) bool {
	switch tableName {
	case "red_rain", "player_red_rain":
		return true
	default:
		return false
	}
}

func dealPlayerRedRain(values []interface{}) {
	playerRedRainLog := &PlayerRedRainsLogs{}
	if err := redis.ScanStruct(values, playerRedRainLog); err != nil {
		logrus.Errorf("redis scan struct error %v", err)
		return
	}
	if playerRedRainLog == nil {
		return
	}
	UpOrCreatePlayerRedRainLog(playerRedRainLog)
}

func dealRedRain(values []interface{}) {
	redRain := &RedRains{}
	if err := redis.ScanStruct(values, redRain); err != nil {
		logrus.Errorf("redis scan struct error %v", err)
		return
	}
	if redRain == nil {
		return
	}
	UpOrCreateRedRain(redRain)
}

func UpOrCreateRedRain(redRain *RedRains) {
	tmpRedRain := &RedRains{Id: redRain.Id}
	err := GetRedRain(tmpRedRain)
	if err == gorm.ErrRecordNotFound {
		if err := CreateRedRain(redRain); err != nil {
			logrus.Errorf("UpOrCreateRedRain  create error %v  %v", err, redRain)
		}
		return
	}
	if err != nil {
		logrus.Errorf("UpOrCreateRedRain get error %v  %v", err, redRain)
		return
	}
	UpRedRain(redRain)
}

func UpOrCreatePlayerRedRainLog(playerRedRainLog *PlayerRedRainsLogs) {
	tmpLog := &PlayerRedRainsLogs{Id: playerRedRainLog.Id}
	err := GetPlayerRedRainLog(tmpLog)
	if err == gorm.ErrRecordNotFound {
		if err := CreatePlayerRedRainLog(playerRedRainLog); err != nil {
			logrus.Errorf("UpOrCreatePlayerRedRainLog  create error %v  %v", err, playerRedRainLog)
		}
		return
	}
	if err != nil {
		logrus.Errorf("UpOrCreatePlayerRedRainLog get error %v  %v", err, playerRedRainLog)
		return
	}
	UpPlayerRedRainLog(playerRedRainLog)
}
