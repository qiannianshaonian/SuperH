package api

import "SuperH/proto"

func InitApi() {
	setRoomProto()
}

func setRoomProto() {
	proto.SetSubInfo("login", "hanz/login/auth", "superhz/login/auth/", GetRoomListApi)
}
