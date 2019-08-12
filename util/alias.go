package util

import (
	"bytes"
	"encoding/base64"
	"github.com/hetianyi/gox/convert"
	"math/rand"
	"time"
)

var (
	rander *rand.Rand
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
	buff.WriteString(string(bs[5:]))
	buff.WriteString("|")
	buff.WriteString(convert.IntToStr(CreateRandNumber(100)))
	return base64.StdEncoding.EncodeToString(buff.Bytes())
}
