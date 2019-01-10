package bridgev2

import (
	"util/json"
	"errors"
)

const (
	// frame head
	FRAME_HEAD_FLAG byte = 66

)

type Frame struct {
	frameHead []byte
	frameStatus byte
	metaLength int
	bodyLength int64
	frameMeta []byte
}


func (frame *Frame) SetOperation(operation byte) {
	frame.frameHead = []byte{FRAME_HEAD_FLAG, operation}
}

func (frame *Frame) GetOperation() byte {
	if frame.frameHead == nil || len(frame.frameHead) != 2 {
		return FRAME_OPERATION_NONE
	}
	return frame.frameHead[1]
}

func (frame *Frame) SetStatus(status byte) {
	frame.frameStatus = status
}

func (frame *Frame) GetStatus() byte {
	return frame.frameStatus
}



func (frame *Frame) SetMeta(meta interface{}) error {
	if meta == nil {
		return errors.New("cannot set frame meta to nil")
	}
	bs, e := json.Marshal(&meta)
	if e != nil {
		return e
	}
	frame.frameMeta = bs
	frame.metaLength = len(bs)
	return nil
}

func (frame *Frame) GetMeta() ([]byte) {
	return frame.frameMeta
}


func (frame *Frame) SetMetaBodyLength(bodyLength int64)  {
	frame.bodyLength = bodyLength
}







