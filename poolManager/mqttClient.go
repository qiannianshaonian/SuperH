package poolManager

import (
	"SuperH/config"
	"SuperH/proto"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"strings"
	"sync/atomic"
	"time"
)

var (
	pubChanMap        map[int64]chan interface{}
	pubChanRoundIndex int64
)

var mqttCallBack mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	noticeCal(msg)
}

func getMqttOpts(ClientId string) *mqtt.ClientOptions { //config "121.41.33.110:1883"
	viper, _ := config.Instance()
	host := viper.GetString("MQTT.host")
	port := viper.GetString("MQTT.port")
	url := fmt.Sprintf("%s:%s", host, port)
	opts := mqtt.NewClientOptions().AddBroker(url).SetClientID(ClientId)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetConnectTimeout(60 * time.Second)
	opts.SetProtocolVersion(4)
	opts.SetAutoReconnect(true)
	return opts
}

func ConnectMqttPubs() {
	pubChanMap = make(map[int64]chan interface{})
	for i := int64(0); i < 10; i++ {
		ch := make(chan interface{}, 3000)
		pubChanMap[i] = ch
		go ConnectMqttPub(i)
	}
}

func ConnectMqttPub(index int64) {
	clientId := fmt.Sprintf("han-pub%d", index)
	opts := getMqttOpts(clientId)
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	for {
		select {
		case t, ok := <-pubChanMap[index]:
			if !ok {
				return
			}
			c.Publish(t.(*proto.GameMsg).PubTopic, 0, false, t.(*proto.GameMsg).Msg)
		}
	}
}

func PubMsg(msg *proto.GameMsg) {
	if len(pubChanMap) < 1 {
		logrus.Errorf("error pub chan map null")
		return
	}
	index := atomic.AddInt64(&pubChanRoundIndex, 1)
	if index > 10000000 {
		atomic.StoreInt64(&pubChanRoundIndex, 0)
	}
	newIndex := index % int64(len(pubChanMap))
	var err error
	msg.Msg, err = json.Marshal(msg.Msg)
	if err != nil {
		logrus.Errorf("error json marshal %v", err)
		return
	}
	pubChanMap[newIndex] <- msg
}

func InitSub() {
	for key, subs := range proto.SubMap {
		if key == "room" {
			for topic, _ := range subs {
				go connectMqttSub(topic)
			}
		} else {
			go connectMqttSub(key)
		}
	}
}

func connectMqttSub(clientId string) {
	opts := getMqttOpts(clientId)
	opts.SetDefaultPublishHandler(mqttCallBack)
	opts.SetOnConnectHandler(onConnectCallBack)
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	for {
		select {}
	}
}

var onConnectCallBack mqtt.OnConnectHandler = func(client mqtt.Client) {
	options := client.OptionsReader()
	clientId := options.ClientID()
	strs := strings.Split(clientId, "_")
	for key, topicMap := range proto.SubMap {
		if len(strs) > 1 {
			for topic, _ := range topicMap {
				if topic == clientId {
					if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
						logrus.Errorf("sub error %v", token.Error())
					}
				}
			}
			continue
		}
		if key != clientId {
			continue
		}
		for topic, _ := range topicMap {
			if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
				logrus.Errorf("sub error %v", token.Error())
			}
		}
	}
}
