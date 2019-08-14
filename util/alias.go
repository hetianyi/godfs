package util

import (
	"bytes"
	"encoding/base64"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"math/rand"
	"time"
)

var (
	rander       *rand.Rand
	aesEncDecKey = []byte("s8f1lf6nm-lqe9z6smoiw-2k8d6w4nla")
)

func init() {
	rander = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func CreateRandNumber(max int) int {
	return rander.Intn(max)
}

func CreateAlias(fid string, instanceId string, ts time.Time) string {
	tsBuff := make([]byte, 8)
	bs := convert.Length2Bytes(ts.Unix(), tsBuff)
	var buff bytes.Buffer
	buff.WriteString(fid)
	buff.WriteString("|")
	buff.WriteString(instanceId)
	buff.WriteString("|")
	buff.WriteString(string(bs[4:]))
	buff.WriteString("|")
	buff.WriteString(FixZeros(CreateRandNumber(100), 3))
	result, err := AesEncrypt(buff.Bytes(), aesEncDecKey)
	if err != nil {
		logger.Error("error while creating alias: ", err)
	}
	return base64.StdEncoding.EncodeToString(result)
}

func FixZeros(i int, width int) string {
	is := convert.IntToStr(i)
	l := len(is)
	for i = 0; i < (width - l); i++ {
		is = "0" + is
	}
	return is
}
