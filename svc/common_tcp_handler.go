package svc

import (
	"errors"
	"fmt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/reg"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	json "github.com/json-iterator/go"
	"hash"
	"io"
	"os"
	"sync"
	"time"
)

const (
	MaxConnPerServer uint = 100
	counterLoopSize       = 60
	maxUploadFactor       = 1000 // max uploads per sec of.
)

var (
	clientAPI             api.ClientAPI
	writableBinlogManager binlog.XBinlogManager
	// counting traffic within 1 minutes
	counterLoop         [counterLoopSize]int // 64, about 1 minute.
	counterPos          int
	counterLock         *sync.Mutex
	millionSecPerUpload = float32(1000) / float32(maxUploadFactor)
)

func init() {
	counterLoop = [counterLoopSize]int{}
	counterLock = new(sync.Mutex)
}

// DigestProxyWriter is a writer proxy which can calculate crc and md5 for the stream file.
type DigestProxyWriter struct {
	crcH hash.Hash32
	md5H hash.Hash
	out  io.Writer
}

func (w *DigestProxyWriter) Write(p []byte) (n int, err error) {
	n, err = w.crcH.Write(p)
	if err != nil {
		return n, err
	}
	n, err = w.md5H.Write(p)
	if err != nil {
		return n, err
	}
	return w.out.Write(p)
}

// InitializeClientAPI initializes client API.
func InitializeClientAPI(config *api.Config) {
	clientAPI = api.NewClient()
	clientAPI.SetConfig(config)
}

func authenticationHandler(header *common.Header, secret string) (*common.Header, *common.Instance, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, nil, 0, nil
	}
	s := header.Attributes["secret"]
	if s != secret {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, nil, 0, nil
	}

	var instance *common.Instance
	if common.BootAs == common.BOOT_TRACKER {
		// parse instance info.
		s1 := header.Attributes["instance"]
		if s1 != "" {
			instance = &common.Instance{}
			if err := json.Unmarshal([]byte(s1), instance); err != nil {
				return &common.Header{
					Result: common.ERROR,
					Msg:    err.Error(),
				}, nil, nil, 0, err
			}
			if err := reg.Put(instance); err != nil {
				return &common.Header{
					Result: common.ERROR,
					Msg:    err.Error(),
				}, nil, nil, 0, err
			}
		}
	}

	return &common.Header{
		Result: common.SUCCESS,
		Msg:    "authentication success",
	}, instance, nil, 0, nil
}

func updateFileReferenceCount(path string, value int64) error {
	oldFile, err := file.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer oldFile.Close()

	tailRefBytes := make([]byte, 8)
	if _, err := oldFile.Seek(-4, 2); err != nil {
		return err
	}
	if _, err := io.ReadAtLeast(oldFile, tailRefBytes[4:], 4); err != nil {
		return err
	}
	// must add lock
	count := convert.Bytes2Length(tailRefBytes)
	logger.Debug("file referenced count: ", count)
	count += value
	convert.Length2Bytes(count, tailRefBytes)
	if _, err := oldFile.Seek(-4, 2); err != nil {
		return err
	}
	if _, err := oldFile.Write(tailRefBytes[4:]); err != nil {
		return err
	}
	return nil
}

func seekRead(fullPath string, offset, length int64) (io.Reader, int64, error) {
	if !file.Exists(fullPath) {
		return nil, 0, errors.New("file not found")
	}
	fi, err := file.GetFile(fullPath)
	if err != nil {
		return nil, 0, err
	}
	info, err := fi.Stat()
	if err != nil {
		return nil, 0, err
	}
	if info.Size() < 4 {
		return nil, 0, errors.New("invalid format file")
	}
	if offset >= info.Size()-4 {
		offset = info.Size() - 4
	}
	if length == -1 || offset+length >= info.Size()-4 {
		length = info.Size() - 4 - offset
	}
	if _, err := fi.Seek(offset, 0); err != nil {
		return nil, 0, err
	}
	return io.LimitReader(fi, length), length, nil
}

func increaseCountForTheSecond() {
	counterLock.Lock()
	defer counterLock.Unlock()

	counterLoop[counterPos] = counterLoop[counterPos] + 1
}

func sumCounter() int {
	counterLock.Lock()
	defer counterLock.Unlock()

	ret := 0
	for _, v := range counterLoop {
		ret += v
	}
	return ret
}

func startCounterLoop() {
	timer.Start(0, 0, time.Second, func(t *timer.Timer) {
		counterLock.Lock()
		defer counterLock.Unlock()

		counterPos++
		if counterPos > counterLoopSize-1 {
			counterPos = 0
		}
		counterLoop[counterPos] = 0
	})
}

func limit() {
	s := sumCounter()
	if s > 0 {
		a := int(float32(s) / float32(counterLoopSize) * millionSecPerUpload)
		fmt.Println(s, "sleep ", a, "ms")
		time.Sleep(time.Millisecond * time.Duration(a))
	}
}
