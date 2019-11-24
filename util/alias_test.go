package util_test

import (
	"encoding/base64"
	"fmt"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"os"
	"testing"
	"time"
)

func TestCreateAlias(t *testing.T) {
	util.GenerateDecKey("123456")
	now := time.Unix(1574600316, 253740900)
	base64S := util.CreateAlias("G01/64/22/e92c1c72e7fff2801c7d4af5b154f88d", "43f01e05", true, now)
	fmt.Println(base64S)
	fmt.Println(time.Now().Unix())

	fmt.Println(util.ParseAlias(base64S, "123456"))

	buff := make([]byte, 8)
	convert.Length2Bytes(2000000, buff)
	fmt.Println(buff)
	fmt.Println("group1/M00/00/0C/rBNM4lrgBU6AH-5BAAzodQCbVVc333")
	fmt.Println(byte('\n'))
	// group1/M00/00/0C/rBNM4lrgBU6AH-5BAAzodQCbVVc333
	// RzAxLzAwL0UyLzAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OWVhfGFjMzM0M2FjfFEu9Hw2MQ==
	// A3AE1i_kNI5gneeop4tWUocv9bYLwyiXDuJSker1VmeWWJ0ioeLA6jIWyPrtRmsZYarNIt0HprjZgXr0H1QNpg
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
	util.GenerateDecKey("123456")
	fmt.Println(util.ParseAlias("tqZbw_9VNFpaVDaEbcdal8jINLjKQfNN58f9BrfEkKeTEaWObbtYa49jbGfmQilu8r9imVKFdPv5tEo1pjai5g", ""))
	a := []byte{1, 2, 3}
	fmt.Println(a[0:1])
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

func TestN(t *testing.T) {
	//aesEncDecKey := []byte("qmwn-ebrv_tcyx#uzia@ospd!lfkg+jh")
	//						  e10adc3949ba59abbe56e057f20f883e
	//secret := []byte("123456")
	fmt.Println(gox.Md5Sum("123456"))
	var a map[string]int = nil
	fmt.Println(len(a))
}
