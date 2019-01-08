package bridgev2

import (
	"io"
	"util/logger"
)

const (
	// frame head
	FRAME_HEAD_FLAG byte = 66

)

type Frame struct {
	frameHead [2]byte
	metaLength int64
	bodyLength int64
	frameMeta *interface{}
	frameBody *io.Reader
}


func (frame *Frame) SetOperation(operation byte) {
	frame.frameHead = [2]byte{FRAME_HEAD_FLAG, operation}
}

func (frame *Frame) SetMeta(meta *interface{}) {
	if meta == nil {
		logger.Error("cannot set frame meta to nil")
		return
	}
	frame.frameMeta = meta
}








