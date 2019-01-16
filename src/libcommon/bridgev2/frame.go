package bridgev2

import (
	"github.com/json-iterator/go"
	"errors"
)

const (
	// frame head
	FRAME_HEAD_FLAG byte = 66

)

type Frame struct {
	FrameHead []byte
	FrameStatus byte
	MetaLength int
	BodyLength int64
	FrameMeta []byte
	BodyWriterHandler func(manager *ConnectionManager, frame *Frame) error
}


func (frame *Frame) SetOperation(operation byte) {
	frame.FrameHead = []byte{FRAME_HEAD_FLAG, operation}
}

func (frame *Frame) GetOperation() byte {
	if frame.FrameHead == nil || len(frame.FrameHead) != 2 {
		return FRAME_OPERATION_NONE
	}
	return frame.FrameHead[1]
}

func (frame *Frame) SetStatus(status byte) {
	frame.FrameStatus = status
}

func (frame *Frame) GetStatus() byte {
	return frame.FrameStatus
}



func (frame *Frame) SetMeta(meta interface{}) error {
	if meta == nil {
		return errors.New("cannot set frame meta to nil")
	}
	bs, e := json.Marshal(&meta)
	if e != nil {
		return e
	}
	frame.FrameMeta = bs
	frame.MetaLength = len(bs)
	return nil
}

func (frame *Frame) GetMeta() ([]byte) {
	return frame.FrameMeta
}


func (frame *Frame) SetMetaBodyLength(bodyLength int64)  {
	frame.BodyLength = bodyLength
}

