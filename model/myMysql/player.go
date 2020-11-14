package myMysql

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

//player_red_rains_logs
type PlayerRedRainsLogs struct {
	Id        int64 `gorm:"primary_key" json:"id"`
	UserId    int64 `gorm:"not null; default:0; type:int;index;comment:'玩家id'" json:"user_id"`
	RedRainId int64 `gorm:"not null; default:0; type:int;comment:'红包雨ID'" json:"red_rain_id"`
	GoldNum   int64 `gorm:"not null; default:0; type:int;comment:'金币数量'" json:"gold_num"`
	ClickNum  int64 `gorm:"not null; default:0; type:int;comment:'该红包点击数'" json:"click_num"`
	UpdateAt  int64 `gorm:"not null; default:0; type:int;comment:'更新时间'" json:"update_at"`
	CreateAt  int64 `gorm:"not null; default:0; type:int;comment:'创建时间'" json:"create_at"`
}

func CreatePlayerRedRainLog(playerLog *PlayerRedRainsLogs) error {
	return myDb.Create(&playerLog).Error
}

func GetPlayerRedRainLog(playerLog *PlayerRedRainsLogs) error {
	return myDb.First(&playerLog, "user_id=? and red_rain_id=?", playerLog.UserId, playerLog.RedRainId).Error
}

func UpPlayerRedRainLog(playerLog *PlayerRedRainsLogs) bool {
	err := myDb.Table("player_red_rains_logs").Where("id=?", playerLog.Id).Updates(playerLog).Error
	if err != nil {
		logrus.Errorf("up player red_rains error %v id=%v", err, playerLog.Id)
		return false
	}
	return true
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func BranchInsert(objs []interface{}, tableName string) error {
	if len(objs) == 0 {
		return nil
	}
	fieldName := ""
	var valueTypeList []string //reflect.ValueOf(
	from := indirect(reflect.ValueOf(objs[0]))

	fieldT := indirectType(from.Type())
	fieldNum := fieldT.NumField()
	for a := 0; a < fieldNum; a++ {
		name := fieldT.Field(a).Tag.Get("json")
		// 添加字段名
		if a == fieldNum-1 {
			fieldName += fmt.Sprintf("`%s`", name)
		} else {
			fieldName += fmt.Sprintf("`%s`,", name)
		}
		// 获取字段类型
		if fieldT.Field(a).Type.Name() == "string" {
			valueTypeList = append(valueTypeList, "string")
		} else if strings.Index(fieldT.Field(a).Type.Name(), "int64") != -1 {
			valueTypeList = append(valueTypeList, "int64")
		} else if strings.Index(fieldT.Field(a).Type.Name(), "int") != -1 {
			valueTypeList = append(valueTypeList, "int")
		} else if strings.Index(fieldT.Field(a).Type.Name(), "bool") != -1 {
			valueTypeList = append(valueTypeList, "bool")
		}
	}
	var valueList []string
	for _, obj := range objs {
		objV := indirect(reflect.ValueOf(obj))
		v := "("
		for index, i := range valueTypeList {
			if index == fieldNum-1 {
				v += GetFormatFeild(objV, index, i, "")
			} else {
				v += GetFormatFeild(objV, index, i, ",")
			}
		}
		v += ")"
		valueList = append(valueList, v)
	}
	insertSql := fmt.Sprintf("insert into `%s` (%s) values %s", tableName, fieldName, strings.Join(valueList, ",")+";")
	err := myDb.Exec(insertSql).Error
	return err
}

// GetFormatFeild 获取字段类型值转为字符串
func GetFormatFeild(objV reflect.Value, index int, t string, sep string) string {
	v := ""
	if t == "string" {
		v += fmt.Sprintf("'%s'%s", objV.Field(index).String(), sep)
	} else if t == "int64" {
		v += fmt.Sprintf("%d%s", objV.Field(index).Int(), sep)
	} else if t == "int" {
		v += fmt.Sprintf("%d%s", objV.Field(index).Int(), sep)
	} else if t == "bool" {
		v += fmt.Sprintf("%v%s", objV.Field(index).Bool(), sep)
	}
	return v
}
