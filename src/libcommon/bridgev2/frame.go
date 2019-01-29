package bridgev2

import (
	"errors"
	json "github.com/json-iterator/go"
)

const (
	// FrameHeadFlag frame head
	FrameHeadFlag byte = 66
)

// Frame is a tcp frame
type Frame struct {
	FrameHead         []byte
	FrameStatus       byte
	MetaLength        int
	BodyLength        int64
	FrameMeta         []byte
	BodyWriterHandler func(manager *ConnectionManager, frame *Frame) error
}

// SetOperation set operation code of the frame
func (frame *Frame) SetOperation(operation byte) {
	frame.FrameHead = []byte{FrameHeadFlag, operation}
}

// GetOperation get operation code of the frame
func (frame *Frame) GetOperation() byte {
	if frame.FrameHead == nil || len(frame.FrameHead) != 2 {
		return FrameOperationNone
	}
	return frame.FrameHead[1]
}

// SetStatus set frame's status
func (frame *Frame) SetStatus(status byte) {
	frame.FrameStatus = status
}

// GetStatus get frame's status
func (frame *Frame) GetStatus() byte {
	return frame.FrameStatus
}

// SetMeta set frame's meta info
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

// GetMeta get frame's meta info
func (frame *Frame) GetMeta() []byte {
	return frame.FrameMeta
}

// SetMetaBodyLength set frame's body length
func (frame *Frame) SetMetaBodyLength(bodyLength int64) {
	frame.BodyLength = bodyLength
}
