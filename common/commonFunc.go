package common

import (
	"SuperH/proto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

//人均一分钱不在此考虑范围内
func RedPaper(count, money int64) ([]int64, int64, int64) {
	if 2*count > money {
		logrus.Errorf("red paper money not enough error money =%d,count=%d", money, count)
		return nil, 0, 0
	}
	moneyList := make([]int64, 0, count)
	sum := int64(0)
	paperMoney := int64(0)
	maxMoneyIndex := int64(0)
	minMoneyIndex := int64(0)
	tmpMoney := money

	for i := int64(0); i < count; i++ {
		tmpMoney = tmpMoney - paperMoney
		paperMoney = doubleAverage(count-i, tmpMoney)
		moneyList = append(moneyList, paperMoney)
		sum += paperMoney
		if moneyList[maxMoneyIndex] <= paperMoney {
			maxMoneyIndex = i
		}
		if moneyList[minMoneyIndex] >= paperMoney {
			minMoneyIndex = i
		}
	}
	if maxMoneyIndex == minMoneyIndex && moneyList[maxMoneyIndex] > 1 { //所有人钱数一样
		moneyList[maxMoneyIndex]++
		moneyList[0]--
	} else {
		moneyList[maxMoneyIndex]++
		moneyList[minMoneyIndex]--
	}
	if sum != money {
		logrus.Errorf("red paper error money =%d,sum=%d", money, sum)
		return nil, 0, 0
	}
	maxMoneyIndex, minMoneyIndex = randSlice(moneyList)
	return moneyList, maxMoneyIndex, minMoneyIndex
}

func randSlice(list []int64) (int64, int64) {
	j := 0
	for key, _ := range list {
		rand.Seed(time.Now().Unix())
		j = rand.Intn(len(list))
		list[key], list[j] = list[j], list[key]
	}
	MaxMoneyIndex := 0
	MinMoneyIndex := 0

	for i, value := range list {
		if list[MaxMoneyIndex] <= value {
			MaxMoneyIndex = i
		}
		if list[MinMoneyIndex] >= value {
			MinMoneyIndex = i
		}
	}
	return int64(MaxMoneyIndex), int64(MinMoneyIndex)
}

//二倍均值算法,count剩余个数,amount剩余金额
func doubleAverage(count, amount int64) int64 {
	//最小钱
	min := int64(2)
	if count == 1 {
		//返回剩余金额
		return amount
	}
	max := amount - min*count
	avg := max * 2 / count
	if avg <= 0 {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(avg) + min
}

func RemoveDuplicates(a []string) []string {
	ret := make(map[string]bool)
	retSlice := make([]string, 0)
	for _, value := range a {
		if _, ok := ret[value]; !ok {
			ret[value] = true
			retSlice = append(retSlice, value)
		}
	}
	return retSlice
}

func ReadWithSelectStr(ch chan string) string {
	select {
	case x := <-ch:
		return x
	default:
		return ""
	}
}

func GetToken(userId int64, userName, passWord string, publicKey string) (string, error) {
	info := &proto.TokenInfo{
		UserId:   userId,
		UserName: userName,
		PassWord: passWord,
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		logrus.Errorf("json marshall error %v", err)
		return "", err
	}

	token, err := RsaEncrypt(bytes, []byte(publicKey))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}

// DecryptToken 对加密的token进行解码
func DecryptToken(token string, privateKey string) (*proto.TokenInfo, error) {
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		logrus.Errorf("base decode error %v", err)
		return nil, err
	}
	result, err := RsaDecrypt(bytes, []byte(privateKey))
	if err != nil {
		logrus.Errorf("rsa decode error %v", err)
		return nil, err
	}

	info := &proto.TokenInfo{}
	err = json.Unmarshal(result, &info)
	if err != nil {
		logrus.Errorf("json unmashall error %v", err)
		return nil, err
	}
	return info, nil
}

func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
	//解密pem格式的公钥
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, nil
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	return rsa.EncryptPKCS1v15(crand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(ciphertext []byte, privateKey []byte) ([]byte, error) {
	//解密
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	//解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 解密
	return rsa.DecryptPKCS1v15(crand.Reader, priv, ciphertext)
}
