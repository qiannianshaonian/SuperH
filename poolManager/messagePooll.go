package poolManager

import (
	"SuperH/common"
	"SuperH/proto"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strings"
	"time"
)

var (
	calIndex int64
)

func newCal(calGoId string) {
	ch := make(chan interface{}, 30)
	calPoolchan <- calGoId
	calPoolMap.Store(calGoId, ch)
	go calGo(calGoId)
}

func calGo(calGoId string) {
	ch, ok := calPoolMap.Load(calGoId)
	if !ok {
		logrus.Errorf("room not exist error %s", calGoId)
		return
	}
	defer func() {
		calPoolMap.Delete(calGoId)
		close(ch.(chan interface{}))
		strs := strings.Split(calGoId, "_")
		if len(strs) == 2 { //calPool_ 为默认进程需要恢复
			newCal(calGoId)
		}
	}()
	for {
		select {
		case msg, isOk := <-ch.(chan interface{}):
			if !isOk {
				logrus.Errorf("this cal go error %s", calGoId)
				return
			}
			dealMqttMsg(msg)
			strs := strings.Split(calGoId, "_")
			if len(strs) == 2 { //calPool_ 为默认进程
				if "calPool_0" != calGoId {
					calPoolchan <- calGoId
				}
			} else {
				return
			}
		}
	}
}

func dealMqttMsg(msg interface{}) {
	newMsg := msg.(mqtt.Message)
	infos := strings.Split(newMsg.Topic(), "/")
	if len(infos) < 2 {
		return
	}
	moduleInfo, ok := proto.SubMap[infos[1]]
	if !ok {
		logrus.Errorf("error moudle map not exist sub topic %v", newMsg.Topic())
		return
	}
	subInfo, ok := moduleInfo[newMsg.Topic()]
	if !ok {
		logrus.Errorf("error sub map not exist sub topic %v", newMsg.Topic())
		return
	}
	time1 := time.Now().UnixNano()
	rspMsg, err := subInfo.Function(newMsg, subInfo.PubTopic)
	allTime := (time.Now().UnixNano() - time1) / 1000000
	if allTime > 100 {
		logrus.Infof("deal topic........... %v time =%v ms", newMsg.Topic(), allTime)
	}
	if err != nil {
		logrus.Errorf("mqttCallBack error %v %v", subInfo.PubTopic, msg)
		return
	}
	if rspMsg != nil {
		PubMsg(rspMsg)
	}
}

func noticeCal(msg mqtt.Message) {
	calChanId := ""
	switch msg.Topic() {
	case "sugar/room/create_room":
		calChanId = "calPool_0"
	default:
		calChanId = common.ReadWithSelectStr(calPoolchan)
	}
	if calChanId == "" {
		nowtime := time.Now().UnixNano()
		rand.Seed(nowtime)
		calChanId = fmt.Sprintf("tmp_calpool_%d_%d", time.Now().UnixNano(), rand.Int63n(nowtime))
		newCal(calChanId)
	}
	if calChann, ok := calPoolMap.Load(calChanId); ok {
		calChann.(chan interface{}) <- msg
	} else {
		logrus.Errorf("cal chan error not found this %s  %v  %v", calChanId, msg.Payload(), msg.Topic())
	}
}
