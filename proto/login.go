package proto

type AuthReq struct {
	UserName string `json:"user_name"`
	UserPass string `json:"user_pass"`
}

type AuthRep struct {
	Status *RspStatus `json:"status"`
	Token  string     `json:"token"`
}

type LoginReq struct {
	Token string `json:"token"`
}

type ItemInfo struct {
	Id  int64 `json:"id"`
	Num int64 `json:"num"`
}

type LoginRsp struct {
	Status   *RspStatus  `json:"status"`
	NickName string      `json:"nick_name"`
	UserId   int64       `json:"user_id"`
	MaxScore int64       `json:"max_score"`
	GoldNum  int64       `json:"gold_num"`
	ItemList []*ItemInfo `json:"item_list"`
	//汉字布局
}
