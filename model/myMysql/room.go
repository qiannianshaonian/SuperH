package myMysql

import (
	"github.com/sirupsen/logrus"
	"os"
)

type RedRains struct {
	Id          int64  `gorm:"primary_key" json:"id"`
	LogId       int64  `gorm:"not null; default:0; type:int;index;comment:'第三方平台记录ID'" json:"log_id"`
	UserId      int64  `gorm:"not null; default:0; type:int;index;comment:'玩家id'" json:"user_id"`
	RoomId      int64  `gorm:"not null; default:0; type:int;index;comment:'房间ID'" json:"room_id"`
	IsWorld     bool   `gorm:"not null; default:0; type:tinyint;comment:'是否是全服红包'" json:"is_world"` //是否是全服红包
	RoomPicture string `gorm:"not null; default:''; type:varchar(256);comment:'房间头像'" json:"room_picture"`
	Content     string `gorm:"not null; default:''; type:varchar(256);comment:'描述'" json:"content"`
	Avatar      string `gorm:"not null; default:''; type:varchar(256);comment:'头像'" json:"avatar"`
	Nickname    string `gorm:"not null; default:''; type:varchar(256);comment:'姓名'" json:"nickname"`
	GoldNum     int64  `gorm:"not null; default:0; type:int;comment:'红包雨金币数'" json:"gold_num"`
	PackNum     int64  `gorm:"not null; default:0; type:int;comment:'红包数量'" json:"pack_num"`
	Packtime    int64  `gorm:"not null; default:0; type:int;comment:'红包时间'" json:"pack_time"`
	ClickNum    int64  `gorm:"not null; default:0; type:int;comment:'被点击多少次'" json:"click_num"`
	HadPeople   int64  `gorm:"not null; default:0; type:int;comment:'多少玩家参入'" json:"had_people"`
	Status      int64  `gorm:"not null; default:0; type:int;comment:'红包状态'" json:"status"` // 0排队等待 1等待 2 下雨等待 3 下雨中 4 结算等待 5结束
	ChanKey     string `gorm:"not null; default:''; type:varchar(128);comment:'房间chan'" json:"chan_key"`
	StartAt     int64  `gorm:"not null; default:0; type:int;comment:'开始时间'" json:"start_at"`
	CreateAt    int64  `gorm:"not null; default:0; type:int;comment:'创建时间'" json:"create_at"`
}

type AddRedRain struct {
	LogId int64
	Type  int64
	Users string
}

func CreateRedRain(redRains *RedRains) error {
	return myDb.Create(&redRains).Error
}

func GetRedRain(redRains *RedRains) error {
	return myDb.First(&redRains, "id=?", redRains.Id).Error
}

func UpRedRain(redRains *RedRains) bool {
	err := myDb.Table("red_rains").Where("id=?", redRains.Id).Updates(redRains).Error
	if err != nil {
		logrus.Errorf("up red_rains error %v id=%v", err, redRains.Id)
		return false
	}
	return true
}

func GetRedRainIdMax() (maxId int64) {
	err := myDb.Table("red_rains").Select(" ifnull(max(id),0)").Find(&maxId).Error
	if err == nil {
		return
	} else {
		logrus.Errorf("get red_rains max id error %v", err)
		os.Exit(1)
	}
	return 0
}

func GetAllRedRains() []*RedRains {
	list := make([]*RedRains, 0)
	if err := myDb.Table("red_rains").Where("status <> ?", 5).Find(&list).Error; err != nil {
		logrus.Errorf("get all red rain error %v", err)
		return nil
	}
	return list
}
