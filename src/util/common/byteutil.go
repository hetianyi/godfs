package common

import "encoding/binary"

func ConvertLen2Bytes(len int64, buffer *[]byte) *[]byte {
    binary.BigEndian.PutUint64(*buffer, uint64(len))
    return buffer
}
