package myRedis

import (
	"SuperH/common"
	"SuperH/config"
	"SuperH/model/myMysql"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

var redisClient *redis.Pool
var redisToMysqlMap = make(map[string]string)

func InitRedisPool() {
	viper, _ := config.Instance()
	host := viper.GetString("REDIS.host")
	port := viper.GetString("REDIS.port")
	pwd := viper.GetString("REDIS.pwd")
	maxconns := viper.GetInt("REDIS.maxconns")
	idleconns := viper.GetInt("REDIS.idleconns")
	hostPort := host + ":" + port
	redisClient = &redis.Pool{
		MaxIdle:     idleconns,
		MaxActive:   maxconns,
		IdleTimeout: 5 * time.Second,
		Wait:        true,
		Dial: func() (conn redis.Conn, e error) {
			con, err := redis.Dial("tcp", hostPort,
				redis.DialPassword(pwd),
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second),
			)
			if err != nil {
				return nil, err
			}
			return con, nil
		},
	}
	client := redisClient.Get()
	rsp, err := redis.String(client.Do("ping"))
	if err != nil {
		logrus.Errorf("redis conn error.... %v %v", rsp, err)
		os.Exit(1)
	}
	defer client.Close()
	logrus.Infof(".... %v %v", rsp, err)
}

func GetRedisClientPool() *redis.Pool {
	return redisClient
}

func RecoverHmsetAndSetExpire(client redis.Conn, key string, value interface{}, extime string) bool {
	client.Send("hmset", redis.Args{}.Add(key).AddFlat(value)...)
	client.Send("EXPIRE", key, extime)
	client.Flush()
	v1, err := client.Receive()
	_, err1 := client.Receive()
	if err != nil || err1 != nil {
		logrus.Errorf("redis set and set expire room error = %v,error1 =%v value=%v  %v", err, err1, value, v1)
		return false
	}
	return true
}

func HmsetAndSetExpire(client redis.Conn, key string, value interface{}, extime string) bool {
	client.Send("hmset", redis.Args{}.Add(key).AddFlat(value)...)
	client.Send("EXPIRE", key, extime)
	strs := strings.Split(key, ":")
	isPush := false
	if len(strs) > 1 {
		if myMysql.IsNeedUpTable(strs[0]) {
			mysqlKey := getMysqlUpListKey()
			client.Send("RPUSH", mysqlKey, key)
			isPush = true
		}
	}
	client.Flush()
	_, err := client.Receive()
	_, err1 := client.Receive()
	if isPush {
		client.Receive()
	}
	if err != nil || err1 != nil {
		logrus.Errorf("redis set and set expire room error = %v,error1 =%v value=%v ", err, err1, value)
		return false
	}
	return true
}

func DoMysqlUpTask() {
	for i := 0; i < 20; i++ {
		if doMysqlUp(500) < 50 {
			return
		}
	}
}

func doMysqlUp(num int64) int64 {
	client := GetRedisClientPool().Get()
	if client.Err() != nil {
		logrus.Errorf("redis get client error %v", client.Err())
		return 0
	}

	defer client.Close()
	redisKeys := GetMysqlUpKeyList(num, client)
	newRedisKeys := common.RemoveDuplicates(redisKeys)
	for _, redisKey := range newRedisKeys {
		client.Send("hgetall", redisKey)
	}
	client.Flush()
	for _, redisKey := range newRedisKeys {
		value, err := redis.Values(client.Receive())
		if err != nil {
			logrus.Errorf("redis get value error %v %v", err, redisKey)
			continue
		}
		strs := strings.Split(redisKey, ":")
		if len(strs) < 1 {
			logrus.Errorf("redis get value error splite %v %v", redisKey)
			continue
		}
		myMysql.UpMysql(value, strs[0])
	}
	return int64(len(newRedisKeys))
}

func RecoverMysqlUpTask() {
	client := GetRedisClientPool().Get()
	if client.Err() != nil {
		logrus.Errorf("redis get client error %v", client.Err())
		return
	}
	defer client.Close()
	num := GetMysqlNeedUpLen(client)
	if num > 0 {
		doMysqlUp(num)
	}
}

func GetMysqlNeedUpLen(client redis.Conn) int64 {
	key := getMysqlUpListKey()
	num, err := redis.Int64(client.Do("LLEN", key))
	if err != nil {
		logrus.Errorf("redis LLEN error %v", err)
		return 0
	}
	return num
}
