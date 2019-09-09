package util

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"math/rand"
	"strings"
	"time"
)

var (
	rander           *rand.Rand
	aesEncDecKey     = []byte("s8f1lf6nm-lqe9z6smoiw-2k8d6w4nla")
	ErrInvalidFileId = errors.New("invalid file id")
)

func init() {
	rander = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func CreateRandNumber(max int) int {
	return rander.Intn(max)
}

// CreateAlias create an alias name from file meta info.
func CreateAlias(fid string, instanceId string, isPrivate bool, ts time.Time) string {
	tsBuff := make([]byte, 8)
	bs := convert.Length2Bytes(ts.Unix(), tsBuff)
	var buff bytes.Buffer
	buff.WriteString(fid)
	buff.WriteString("|")
	buff.WriteString(instanceId)
	buff.WriteString("|")
	buff.WriteString(gox.TValue(isPrivate, "1", "0").(string))
	buff.WriteString("|")
	buff.WriteString(string(bs[4:]))
	buff.WriteString("|")
	buff.WriteString(FixZeros(CreateRandNumber(100), 3))
	result, err := AesEncrypt(buff.Bytes(), aesEncDecKey)
	if err != nil {
		logger.Error("error while creating alias: ", err)
	}
	return base64.RawURLEncoding.EncodeToString(result)
}

// ParseAlias parses file info from file alias name,
// and returns *common.FileInfo
func ParseAlias(alias string) (fileInfo *common.FileInfo, err error) {
	bs, err := base64.RawURLEncoding.DecodeString(alias)
	if err != nil {
		return
	}
	recovered, err := AesDecrypt(bs, aesEncDecKey)
	if err != nil {
		return
	}
	parts := strings.Split(string(recovered), "|")
	if len(parts) != 5 {
		return nil, errors.New("invalid format fileId")
	}
	if !common.FileMetaPatternRegexp.Match([]byte(parts[0])) || len(parts[1]) != 8 {
		return nil, ErrInvalidFileId
	}

	group := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$1")
	p1 := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$2")
	p2 := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$3")
	md5 := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$4")

	tsBuff := []byte{0, 0, 0, 0, parts[3][0], parts[3][1], parts[3][2], parts[3][3]}
	ts := convert.Bytes2Length(tsBuff)
	return &common.FileInfo{
		Group:      group,
		FileLength: 0,
		Path:       strings.Join([]string{p1, p2, md5}, "/"),
		InstanceId: parts[1],
		IsPrivate:  gox.TValue(parts[2] == "1", true, false).(bool),
		CreateTime: ts,
	}, nil
}

func FixZeros(i int, width int) string {
	is := convert.IntToStr(i)
	l := len(is)
	for i = 0; i < (width - l); i++ {
		is = "0" + is
	}
	return is
}
