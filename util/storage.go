package util

import (
	"container/list"
	"encoding/base64"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
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

// GenerateToken generates token for http file download.
func GenerateToken(fileId, secret string, expireTimestamp string) string {
	return gox.Md5Sum(expireTimestamp, fileId, secret)
}

func ExistsFile(fInfo *common.FileInfo) bool {
	if common.BootAs == common.BOOT_STORAGE {
		return file.Exists(common.InitializedStorageConfiguration.DataDir + "/" + fInfo.Path)
	}
	return false
}

// ClearList clears the list elements.
func ClearList(l *list.List) {
	if l == nil {
		return
	}
	for l.Front() != nil {
		l.Remove(l.Front())
	}
}
