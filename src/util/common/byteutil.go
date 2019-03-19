package common

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io"
)

func ConvertLen2Bytes(len int64, buffer *[]byte) *[]byte {
	binary.BigEndian.PutUint64(*buffer, uint64(len))
	return buffer
}

func Md5sum(input ...string) string {
	h := md5.New()
	if input != nil {
		for _, v := range input {
			io.WriteString(h, v)
		}
	}
	sliceCipherStr := h.Sum(nil)
	sMd5 := hex.EncodeToString(sliceCipherStr)
	return sMd5
}
