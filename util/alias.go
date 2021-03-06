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
	rander               *rand.Rand
	aesEncDecKey         []byte
	ErrInvalidFileId     = errors.New("invalid fileId")
	historyAesEncDecKeys map[string]string
)

func init() {
	rander = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// CreateRandNumber create a random int value.
func CreateRandNumber(max int) int {
	return rander.Intn(max)
}

// AddSecretEncryptKeys adds all history secrets of all storage servers.
func AddSecretEncryptKeys(secret ...string) {
	if secret == nil || len(secret) == 0 {
		return
	}
	if historyAesEncDecKeys == nil {
		historyAesEncDecKeys = make(map[string]string)
	}
	for _, k := range secret {
		historyAesEncDecKeys[gox.Md5Sum(k)] = k
	}
}

// GenerateDecKey generates aes encrypt key for current usage.
func GenerateDecKey(secret string) {
	aesEncDecKey = []byte(gox.Md5Sum(secret))
}

// CreateAlias create an 86 bytes long, self description alias name from file meta info.
//
// Alias name contains file's path, instanceId, private flag, timestamp and a random number,
// these information combines with '|' and encrypt using AES.
//
// Alias name is a little bit long...
func CreateAlias(fid string, instanceId string, isPrivate bool, ts time.Time) string {
	tsBuff := make([]byte, 8)
	bs := convert.Length2Bytes(ts.Unix(), tsBuff)

	var buff bytes.Buffer
	buff.WriteString(fid)
	buff.WriteString(instanceId)
	buff.WriteString(gox.TValue(isPrivate, "1", "0").(string))
	buff.WriteString(string(bs[4:]))
	buff.WriteString(FixZeros(CreateRandNumber(100), 3))

	result, err := AesEncrypt(buff.Bytes(), aesEncDecKey)
	if err != nil {
		logger.Error("error while creating alias: ", err)
	}
	return base64.RawURLEncoding.EncodeToString(result)
}

// ParseAlias parses file info from file alias name,
// and returns *common.FileInfo
func ParseAlias(alias, currentSecret string) (fileInfo *common.FileInfo, secret string, err error) {
	fileInfo, err = parseAliasForSecret(alias, nil)
	secret = currentSecret
	if err == nil || len(historyAesEncDecKeys) == 0 {
		return
	}
	logger.Debug("failed to parse alias by current secret, trying history secrets.")
	for k, v := range historyAesEncDecKeys {
		logger.Debug("trying history secrets: ", v)
		fileInfo, err = parseAliasForSecret(alias, []byte(k))
		if err != nil {
			logger.Debug("failed history secrets: ", v)
			continue
		}
		secret = v
		logger.Debug("secrets worked: ", v)
		break
	}
	return
}

// parseAliasForSecret parses fileId from history secrets.
func parseAliasForSecret(alias string, aesKey []byte) (fileInfo *common.FileInfo, err error) {
	gox.Try(func() {
		bs, e := base64.RawURLEncoding.DecodeString(alias)
		if e != nil {
			err = e
			return
		}
		recovered, e := AesDecrypt(bs, gox.TValue(aesKey == nil, aesEncDecKey, aesKey).([]byte))
		if e != nil {
			err = e
			return
		}
		l := len(recovered)
		parts := []string{
			string(recovered[:l-16]),
			string(recovered[l-16 : l-8]),
			string(recovered[l-8 : l-7]),
			string(recovered[l-7 : l-3]),
			string(recovered[l-3 : l]),
		}
		if !common.FileMetaPatternRegexp.Match([]byte(parts[0])) || len(parts[1]) != 8 {
			err = ErrInvalidFileId
			return
		}

		group := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$1")
		p1 := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$2")
		p2 := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$3")
		md5 := common.FileMetaPatternRegexp.ReplaceAllString(parts[0], "$4")

		tsBuff := []byte{0, 0, 0, 0, parts[3][0], parts[3][1], parts[3][2], parts[3][3]}
		ts := convert.Bytes2Length(tsBuff)
		fileInfo = &common.FileInfo{
			Group:      group,
			FileLength: 0,
			Path:       strings.Join([]string{p1, p2, md5}, "/"),
			InstanceId: parts[1],
			IsPrivate:  gox.TValue(parts[2] == "1", true, false).(bool),
			CreateTime: ts,
		}
		return
	}, func(e interface{}) {
		err = ErrInvalidFileId
		logger.Debug("error parsing fileId: ", e, " for ", alias)
	})
	return
}

// FixZeros prepends '0's to a int value.
func FixZeros(i int, width int) string {
	is := convert.IntToStr(i)
	l := len(is)
	for i = 0; i < (width - l); i++ {
		is = "0" + is
	}
	return is
}
