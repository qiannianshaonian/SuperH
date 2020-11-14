package myRedis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"math/rand"
	"red-envelope-rain/model/myMysql"
	"red-envelope-rain/proto"
	"strconv"
	"time"
)

func UPRedRain(rains *myMysql.RedRains, client redis.Conn, exTime int64) bool {
	key := GetRedRainKey(rains.Id)
	if !HmsetAndSetExpire(client, key, rains, strconv.FormatInt(exTime, 10)) {
		return false
	}
	return true
}

func PushAddUserMoneyQueue(addRedRain *myMysql.AddRedRain, client redis.Conn) bool {
	if client == nil {
		client = GetRedisClientPool().Get()
		if client.Err() != nil {
			logrus.Errorf("redis get client error %v", client.Err())
			return false
		}
		defer client.Close()
	}
	key := GetAddUserMoneyKey(addRedRain.LogId)
	queueKey := getAddUserMoneyQueueKey()
	rand.Seed(time.Now().Unix())
	client.Send("hmset", redis.Args{}.Add(key).AddFlat(addRedRain)...)
	client.Send("EXPIRE", key, 86400+rand.Int63n(1000))
	client.Send("RPUSH", queueKey, key)
	client.Flush()
	_, err := client.Receive()
	_, err1 := client.Receive()

	_, err2 := client.Receive()
	if err != nil || err1 != nil || err2 != nil {
		logrus.Errorf("redis set and set expire add user money error = %v,error1 =%v err2=%v ", err, err1, err2)
		return false
	}
	return true
}

func GetAddUserMoneyKeyList(num int64, client redis.Conn) []string {
	key := getAddUserMoneyQueueKey()
	redisKeys, err := redis.Strings(client.Do("lrange", key, 0, num))
	if err != nil {
		logrus.Errorf("redis lrange error %v", err)
		return nil
	}

	client.Do("LTRIM", key, len(redisKeys), -1)
	return redisKeys
}

func GetRedRain(radRainId int64, client redis.Conn) *myMysql.RedRains {
	key := GetRedRainKey(radRainId)
	redRain, err := redis.Values(client.Do("hgetall", key))
	if err != nil {
		logrus.Errorf("redis hget error %v userid=%v", err, key)
		return nil
	}
	// 需要注意频繁不存在的 频繁查mysql问题
	redRainInfo := &myMysql.RedRains{}
	if len(redRain) < 1 {
		return nil
	}
	err = redis.ScanStruct(redRain, redRainInfo)
	if err != nil {
		logrus.Errorf("redis data error %v", err)
		return nil
	}
	return redRainInfo
}

func AddRedRain(redRainId, roomId int64, client redis.Conn) bool {
	key := getRoomRedRainsKey(roomId)
	if _, err := client.Do("sadd", key, redRainId); err != nil {
		logrus.Errorf("redis sadd error %v", err)
		return false
	}
	return true
}

func AddWorldRedRain(redRainId int64, client redis.Conn) bool {
	key := getWorldRedRain()
	if _, err := client.Do("sadd", key, redRainId); err != nil {
		logrus.Errorf("redis sadd error %v", err)
		return false
	}
	return true
}

func RemRedRain(redRainId, roomId int64, client redis.Conn) bool {
	if client == nil {
		client = GetRedisClientPool().Get()
		if client.Err() != nil {
			logrus.Errorf("redis get client error %v", client.Err())
			return false
		}
		defer client.Close()
	}
	key := getRoomRedRainsKey(roomId)
	worldKey := getWorldRedRain()
	if _, err := client.Do("srem", key, redRainId); err != nil {
		logrus.Errorf("redis srem error %v", err)
		return false
	}
	if _, err := client.Do("srem", worldKey, redRainId); err != nil {
		logrus.Errorf("redis world srem error %v", err)
		return false
	}
	return true
}

func GetRedRains(roomId int64, client redis.Conn) []int64 {
	key := getRoomRedRainsKey(roomId)
	redRainIds, err := redis.Int64s(client.Do("SMEMBERS", key))
	if err != nil {
		return nil
	}
	return redRainIds
}

func GetWorldRedRains(client redis.Conn) []int64 {
	key := getWorldRedRain()
	redRainIds, err := redis.Int64s(client.Do("SMEMBERS", key))
	if err != nil {
		return nil
	}
	return redRainIds
}

func GetRoomChan(roomId int64, client redis.Conn) string {
	key := getRoomChanKey(roomId)
	roomChanKey, err := redis.String(client.Do("get", key))
	if err != nil {
		return ""
	}
	return roomChanKey
}

func UpRoomChan(roomId int64, chanKey string, client redis.Conn) bool {
	key := getRoomChanKey(roomId)
	if _, err := client.Do("set", key, chanKey, "EX", strconv.FormatInt(86400, 10)); err != nil {
		logrus.Errorf("redis set error %v", err)
		return false
	}
	return true
}

func GetMysqlUpKeyList(num int64, client redis.Conn) []string {
	key := getMysqlUpListKey()
	redisKeys, err := redis.Strings(client.Do("lrange", key, 0, num))
	if err != nil {
		logrus.Errorf("redis lrange error %v", err)
		return nil
	}

	client.Do("LTRIM", key, len(redisKeys), -1)
	return redisKeys
}

func GetUserBase(token string, client redis.Conn) *proto.UserData {
	if client == nil {
		client = GetRedisClientPool().Get()
		if client.Err() != nil {
			logrus.Errorf("redis get client error %v", client.Err())
			return nil
		}
		defer client.Close()
	}
	key := getUserInfoKey(token)
	userDatas, err := redis.Values(client.Do("hgetall", key))
	if err != nil {
		logrus.Errorf("redis hget error %v userid=%v", err, key)
		return nil
	}
	// 需要注意频繁不存在的 频繁查mysql问题
	userData := &proto.UserData{}
	if len(userDatas) < 1 {
		return nil
	}
	err = redis.ScanStruct(userDatas, userData)
	if err != nil {
		logrus.Errorf("redis data error %v", err)
		return nil
	}
	return userData
}

func UpUserData(userData *proto.UserData, client redis.Conn) bool {
	if client == nil {
		client = GetRedisClientPool().Get()
		if client.Err() != nil {
			logrus.Errorf("redis get client error %v", client.Err())
			return false
		}
		defer client.Close()
	}
	key := getUserInfoKey(userData.Token)
	rand.Seed(time.Now().Unix())
	if !HmsetAndSetExpire(client, key, userData, strconv.FormatInt(3600+rand.Int63n(100), 10)) {
		return false
	}
	return true
}

func GetRedRainKey(redRainId int64) string {
	return fmt.Sprintf("red_rain:red_rain_id:%d", redRainId)
}

func getPlayerRedRainKey(userId, redRainId int64) string {
	return fmt.Sprintf("player_red_rain:red_rain_id:%d:user_id:%d", redRainId, userId)
}

func getRedRainUsersKey(redRainId int64) string {
	return fmt.Sprintf("red_rain_users:red_rain_id:%d", redRainId)
}

func getRoomRedRainsKey(roomId int64) string {
	return fmt.Sprintf("room_red_rains:room_id:%d", roomId)
}

func getMysqlUpListKey() string {
	return "mysql_up_list"
}

func getRoomChanKey(roomId int64) string {
	return fmt.Sprintf("room_chan:room_id:%d", roomId)
}

func getWorldRedRain() string {
	return fmt.Sprintf("world_red_rains")
}

func getUserInfoKey(token string) string {
	return fmt.Sprintf("user_base:token:%s", token)
}

func GetAddUserMoneyKey(logId int64) string {
	return fmt.Sprintf("add_user_moneys:%d", logId)
}

func getAddUserMoneyQueueKey() string {
	return fmt.Sprintf("add_user_money_queue")
}
