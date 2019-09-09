package util_test

import (
	"encoding/base64"
	"fmt"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"os"
	"testing"
	"time"
)

func TestCreateAlias(t *testing.T) {
	base64S := util.CreateAlias("G01/00/E2/012345678901234567890123456789ea", "ac3343ac", true, time.Now())
	fmt.Println(base64S)
	fmt.Println(time.Now().Unix())

	buff := make([]byte, 8)
	convert.Length2Bytes(2000000, buff)
	fmt.Println(buff)
	fmt.Println(byte('\n'))
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

func TestParseAlias(t *testing.T) {
	fmt.Println(util.ParseAlias("A3AE1i_kNI5gneeop4tWUocv9bYLwyiXDuJSker1VmeWWJ0ioeLA6jIWyPrtRmsZ_RBn0tWAeRXQ8o3lnqWjpg"))
}

func TestSeek(t *testing.T) {
	a := "D:/seek"
	fi, err := file.OpenFile(a, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fi.Close()
	fi.WriteAt([]byte{0}, 2999)
}
