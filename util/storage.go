package util

import (
	"encoding/base64"
	"fmt"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"math/rand"
	"time"
)

// CreateCRCFileID creates a file id by instance, crc, file length, timestamp and random int
func CreateCRCFileID(instanceId string, crc32 string, fileSize uint64) string {
	timestamp := fmt.Sprintf("%x", gox.GetTimestamp(time.Now()))
	fileSizeHex := fmt.Sprintf("%x", fileSize)
	randInt := ""
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 3; i++ {
		randInt += convert.IntToStr(rnd.Intn(10))
	}
	all := instanceId + timestamp + crc32 + fileSizeHex
	return fmt.Sprintf("%s%s", base64.StdEncoding.EncodeToString([]byte(all)), randInt)
}

// CreateCRCFileID creates a file id by instance, md5, timestamp and random int
func CreateMD5FileID(instanceId string, md5 string) string {
	timestamp := fmt.Sprintf("%x", gox.GetTimestamp(time.Now()))
	randInt := ""
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 3; i++ {
		randInt += convert.IntToStr(rnd.Intn(10))
	}
	all := instanceId + timestamp + md5
	return fmt.Sprintf("%s%s", base64.StdEncoding.EncodeToString([]byte(all)), randInt)
}
