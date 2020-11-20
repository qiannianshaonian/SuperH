package proto

import mqtt "github.com/eclipse/paho.mqtt.golang"

type GameMsg struct {
	PubTopic string
	Msg      interface{}
}

type RspCode int32

type SubInfo struct {
	PubTopic string
	Function func(mqtt.Message, string) (*GameMsg, error)
}

type RspStatus struct {
	Code RspCode `json:"code"`
	Msg  string  `json:"msg"`
}

var (
	SubMap       = make(map[string]map[string]*SubInfo)
	RspStatusMap = make(map[RspCode]string)
)

func SetSubInfo(moduleName, subTopic, pubTopic string, funcs func(mqtt.Message, string) (*GameMsg, error)) {
	if v, ok := SubMap[moduleName]; ok {
		v[subTopic] = &SubInfo{
			PubTopic: pubTopic,
			Function: funcs,
		}
	} else {
		SubMap[moduleName] = make(map[string]*SubInfo)
		SubMap[moduleName][subTopic] = &SubInfo{
			PubTopic: pubTopic,
			Function: funcs,
		}
	}
}

func GetRspStatus(rspStatus *RspStatus, code RspCode) {
	rspStatus.Code = code
	if codeMsg, ok := RspStatusMap[code]; ok {
		rspStatus.Msg = codeMsg
		return
	}
	rspStatus.Msg = ""
}
