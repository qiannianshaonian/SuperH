package control

import "SuperH/proto"

func Auth(req *proto.AuthReq, rsp *proto.AuthRsp) proto.RspCode {

	return proto.Rsp_code_success
}
