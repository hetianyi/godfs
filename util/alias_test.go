package util_test

import (
	"encoding/base64"
	"fmt"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"testing"
	"time"
)

func TestCreateAlias(t *testing.T) {
	base64S := util.CreateAlias("G01/00/E2/012345678901234567890123456789ea", "ac3343ac", time.Now())
	fmt.Println(base64S)
	fmt.Println(time.Now().Unix())

	buff := make([]byte, 8)
	convert.Length2Bytes(1, buff)
	fmt.Println(buff)
	// group1/M00/00/0C/rBNM4lrgBU6AH-5BAAzodQCbVVc333
	// RzAxLzAwL0UyLzAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OWVhfGFjMzM0M2FjfFEu9Hw2MQ==
	// G01/00/E2/MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5ZWF8YWMzMzQzYWN8Ukg9fDAxMA==
}

func TestAesCbcEncrypt(t *testing.T) {
	input := []byte("G01/00/E2/0123456789012345g7890123456789eaac5343ac08")
	key := []byte("1234567890123456")
	bs, err := util.AesEncrypt(input, key)
	if err != nil {
		logger.Error(err)
	} else {
		ret := base64.StdEncoding.EncodeToString(bs)
		fmt.Println(ret)

		bs, _ := base64.StdEncoding.DecodeString(ret)
		recovered, _ := util.AesDecrypt(bs, key)
		fmt.Println(string(recovered))
	}
}
