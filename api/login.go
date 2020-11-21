package api

import (
	"SuperH/control"
	"SuperH/proto"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

func AuthApi(msg mqtt.Message, pubTopic string) (*proto.GameMsg, error) {
	rsp := &proto.AuthRsp{
		Status: &proto.RspStatus{
			Code: proto.Rsp_code_error,
			Msg:  proto.Rsp_code_error_msg,
		},
	}
	req := &proto.AuthReq{}
	if err := json.Unmarshal(msg.Payload(), req); err != nil {
		logrus.Errorf("json un error %v msg.Payload()=%v", err, msg.Payload())
		return nil, err
	}
	control.Auth(req, rsp)
	gameMsg := &proto.GameMsg{}
	gameMsg.Msg = rsp
	return gameMsg, nil
}
