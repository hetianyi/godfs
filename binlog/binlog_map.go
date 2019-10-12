package binlog

import (
	"errors"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"io/ioutil"
	"os"
	"sync"
)

type XBinlogMapManager struct {
	lock      *sync.Mutex
	binlogDir string
	mapFile   *os.File
	buffer    []byte
	memMap    [1000][3]byte
}

func (m *XBinlogMapManager) initMapFile() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// try to close old connection.
	if m.mapFile != nil {
		m.mapFile.Close()
	}

	binlogMapFile := m.binlogDir + "/binlog.map"
	exist := file.Exists(binlogMapFile)
	ret, err := file.OpenFile(binlogMapFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	if !exist {
		logger.Debug("created binlog map file")
		ret.WriteAt([]byte{0}, 2999)
		m.memMap = [1000][3]byte{}
	} else {
		fInfo, err := os.Stat(binlogMapFile)
		if err != nil {
			return err
		}
		if fInfo.Size() != 3000 {
			return errors.New("error binlog map file")
		}

		data, err := ioutil.ReadFile(binlogMapFile)
		if err != nil {
			return err
		}
		if len(data) != 3000 {
			return errors.New("error binlog map file")
		}
		for i := 0; i < 1000; i++ {
			m.memMap[i] = [3]byte{data[i*3], data[i*3+1], data[i*3+2]}
		}
	}
	m.mapFile = ret
	return nil
}

// GetRecords get binlog record size by binlog file index.
func (m *XBinlogMapManager) GetRecords(fileIndex int) (size int, err error) {
	m.lock.Lock()
	defer func() {
		if err != nil {
			m.initMapFile()
		}
		m.lock.Unlock()
	}()

	d3 := m.memMap[fileIndex]
	m.buffer[5] = d3[0]
	m.buffer[6] = d3[1]
	m.buffer[7] = d3[2]

	return int(convert.Bytes2Length(m.buffer)), nil
	/*
		if _, err := m.mapFile.Seek(int64(fileIndex*3), 0); err != nil {
			return 0, err
		}
		if _, err := io.ReadAtLeast(m.mapFile, m.buffer[5:], 3); err != nil {
			return 0, err
		}
		return int(convert.Bytes2Length(m.buffer)), nil*/
}

func (m *XBinlogMapManager) SetRecords(fileIndex int, value int) (err error) {
	m.lock.Lock()
	defer func() {
		if err != nil {
			m.initMapFile()
		}
		m.lock.Unlock()
	}()

	if _, err := m.mapFile.Seek(int64(fileIndex*3), 0); err != nil {
		return err
	}
	convert.Length2Bytes(int64(value), m.buffer)
	if _, err := m.mapFile.Write(m.buffer[5:]); err != nil {
		return err
	}
	d3 := m.memMap[fileIndex]
	d3[0] = m.buffer[5]
	d3[1] = m.buffer[6]
	d3[2] = m.buffer[7]
	m.memMap[fileIndex] = d3
	return nil
}
