package main

import (
	"SuperH/config"
	"SuperH/logMod"
	"SuperH/model/myMysql"
	"SuperH/model/myRedis"
	"SuperH/poolManager"
)

func main() {
	config.Init()
	logMod.Init()
	//proto.InitRspStatusMap()
	myMysql.InitMysqlDb()
	myRedis.InitRedisPool()
	go poolManager.InitSub()
	poolManager.ConnectMqttPubs()

}
