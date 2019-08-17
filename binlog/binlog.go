package binlog

import (
	"bytes"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"os"
	"sync"
	"time"
)

const (
	LOCAL_BINLOG_MANAGER   XBinlogManagerType = 1
	SYNC_BINLOG_MANAGER    XBinlogManagerType = 2
	TRACKER_BINLOG_MANAGER XBinlogManagerType = 3
	MAX_BINLOG_SIZE        int                = 5 // 200w binlog records
)

var binlogMapManager *XBinlogMapManager

type XBinlogManagerType byte

type XBinlogManager interface {
	// GetType returns this manager type.
	//
	// manager type could be one of:
	// `LOCAL_BINLOG_MANAGER`,
	// `SYNC_BINLOG_MANAGER`,
	// `TRACKER_BINLOG_MANAGER`
	GetType() XBinlogManagerType
	// Write writes a binlog to file.
	Write(bin *common.BingLog) error
	// Read reads binlog from file.
	//
	// fileIndex: the binlog file index, -1 means reads from latest binlog file.
	//
	// offset: read offset in bytes, must be integer multiple of the binlog.
	Read(fileIndex int, offset int64) (*common.BingLog, error)
}

func NewXBinlogManager(managerType XBinlogManagerType) XBinlogManager {
	// check binlog dirs
	binlogDir := getBinlogDir()
	if err := initialBinlogDir(binlogDir); err != nil {
		logger.Fatal("failed to create binlog dir: ", err)
	}
	// initialize XBinlogMapManager
	if binlogMapManager == nil {
		binlogMapManager = &XBinlogMapManager{
			lock:      new(sync.Mutex),
			buffer:    make([]byte, 8),
			binlogDir: binlogDir,
		}
		if err := binlogMapManager.initMapFile(); err != nil {
			logger.Fatal("failed to initialize binlog map file: ", err)
		}
	}
	if managerType == LOCAL_BINLOG_MANAGER {
		return &localBinlogManager{
			writeLock:    new(sync.Mutex),
			binlogSize:   0,
			buffer:       bytes.Buffer{},
			lengthBuffer: make([]byte, 8),
		}
	}
	// TODO
	return nil
}

type localBinlogManager struct {
	writeLock         *sync.Mutex
	currentBinLogFile *os.File // current binlog file
	binlogSize        int      // binlog items count
	buffer            bytes.Buffer
	lengthBuffer      []byte
	currentIndex      int
}

func (m *localBinlogManager) GetType() XBinlogManagerType {
	return LOCAL_BINLOG_MANAGER
}

func (m *localBinlogManager) Write(bin *common.BingLog) error {
	m.writeLock.Lock()
	defer m.writeLock.Unlock()

	logger.Debug("writing binlog")
	if m.currentBinLogFile == nil || m.binlogSize >= MAX_BINLOG_SIZE {
		// file size exceed, close old file
		if m.currentBinLogFile != nil {
			logger.Debug("binlog exceed max size")
			if err := m.currentBinLogFile.Close(); err != nil {
				return err
			}
		}
		// create new binlog file.
		newFile, binLogSize, index, err := getCurrentBinLogFile()
		if err != nil {
			return err
		}
		m.currentBinLogFile = newFile
		m.binlogSize = binLogSize
		m.currentIndex = index
	}

	// build buffer
	defer m.buffer.Reset()
	m.buffer.WriteByte(bin.DownloadFinish)
	m.buffer.Write(bin.SourceInstance[:])
	m.buffer.Write(bin.Timestamp[:])
	m.buffer.Write(bin.FileLength[:])
	m.buffer.Write(bin.FileId[:])
	m.buffer.WriteRune('\n')
	// persist binlog data.
	if _, err := m.currentBinLogFile.Write(m.buffer.Bytes()); err != nil {
		return err
	}
	// sync data.
	if err := m.currentBinLogFile.Sync(); err != nil {
		return err
	}
	m.binlogSize += 1
	// write binlog record size.
	if err := binlogMapManager.SetRecords(m.currentIndex, m.binlogSize); err != nil {
		return err
	}
	logger.Debug("binlog write success")
	return nil
}

func (m *localBinlogManager) Read(fileIndex int, offset int64) (*common.BingLog, error) {
	// TODO
	return nil, nil
}

// Create creates binlog file under datadir.
func create() (*os.File, int, error) {
	logger.Debug("creating binlog file")
	// check binlog dirs
	binlogDir := getBinlogDir()
	if err := initialBinlogDir(binlogDir); err != nil {
		return nil, 0, err
	}
	i := 0
	for {
		binLogFileName := getBinLogFileNameByIndex(binlogDir, i)
		i++
		if file.Exists(binLogFileName) {
			continue
		}
		out, err := file.CreateFile(binLogFileName)
		if err != nil {
			return nil, 0, err
		}
		logger.Debug("binlog file created: ", binLogFileName)
		return out, i - 1, nil
	}
}

// getCurrentBinLogFile gets current binlog file for writing.
//
// Returns the binlog file, binlog record size, binlog file index NO., and error.
func getCurrentBinLogFile() (*os.File, int, int, error) {
	dataDir := ""
	if common.BootAs == common.BOOT_TRACKER {
		dataDir = common.InitializedTrackerConfiguration.DataDir
	} else if common.BootAs == common.BOOT_STORAGE {
		dataDir = common.InitializedStorageConfiguration.DataDir
	} else {
		logger.Fatal("unSupported boot mode: ", common.BootAs)
	}

	// check binlog dirs
	binlogDir := dataDir + "/binlog"
	if !file.Exists(binlogDir) {
		if err := file.CreateDirs(binlogDir); err != nil {
			return nil, 0, 0, err
		}
	}

	latestLogFileName := ""
	shouldExists := false
	index := 0
	for i := 999; i >= 0; i-- {
		name := getBinLogFileNameByIndex(binlogDir, i)
		// find latest binlog file.
		if !shouldExists && file.Exists(name) {
			index = i
			latestLogFileName = name
			shouldExists = true
		}
		// check left binlog file state.
		if shouldExists && !file.Exists(name) {
			return nil, 0, 0, errors.New("invalid binlog state: binlog loss")
		}
	}
	// no binlog file yet.
	if latestLogFileName == "" {
		ret, _index, err := create()
		return ret, 0, _index, err
	}
	latest, err := file.GetFile(latestLogFileName)
	if err != nil {
		return nil, 0, 0, err
	}
	defer latest.Close()
	info, err := latest.Stat()
	if err != nil {
		return nil, 0, 0, err
	}
	if info.IsDir() {
		return nil, 0, 0, errors.New("binlog file must not be a directory: " + info.Name())
	}
	binlogSize, err := binlogMapManager.GetRecords(index)
	if err != nil {
		return nil, 0, 0, err
	}
	// this binlog file exceed max record size.
	if binlogSize >= MAX_BINLOG_SIZE {
		ret, _index, err := create()
		return ret, 0, _index, err
	}
	logger.Debug("use binlog file: ", latestLogFileName)
	ret, err := file.OpenFile(latestLogFileName, os.O_RDWR, 0666)
	return ret, binlogSize, index, err
}

func getBinLogFileNameByIndex(binlogDir string, i int) string {
	if common.BootAs == common.BOOT_TRACKER {
		return binlogDir + "/bin." + util.FixZeros(i, 3)
	} else if common.BootAs == common.BOOT_STORAGE {
		return binlogDir + "/bin." + util.FixZeros(i, 3)
	}
	return ""
}

// CreateLocalBinlog builds an Binlog.
func CreateLocalBinlog(fileId string, fileLength int64, instanceId string, tm time.Time) *common.BingLog {
	buffer8 := make([]byte, 8)
	// file length
	convert.Length2Bytes(fileLength, buffer8)
	var flen = Copy8(buffer8)
	// timestamp
	convert.Length2Bytes(tm.Unix(), buffer8)
	var ts = Copy8(buffer8)
	// instance
	var ins = Copy8([]byte(instanceId))

	return &common.BingLog{
		Type:           byte(LOCAL_BINLOG_MANAGER),
		FileId:         []byte(fileId),
		SourceInstance: ins,
		FileLength:     flen,
		Timestamp:      ts,
		DownloadFinish: 1,
	}
}

func Copy8(src []byte) [8]byte {
	var target [8]byte
	for i := 0; i < 8; i++ {
		target[i] = src[i]
	}
	return target
}

func getBinlogDir() string {
	dataDir := ""
	if common.BootAs == common.BOOT_TRACKER {
		dataDir = common.InitializedTrackerConfiguration.DataDir
	} else if common.BootAs == common.BOOT_STORAGE {
		dataDir = common.InitializedStorageConfiguration.DataDir
	} else {
		logger.Fatal("unSupported boot mode: ", common.BootAs)
	}
	return dataDir + "/binlog"
}

func initialBinlogDir(path string) (err error) {
	if !file.Exists(path) {
		if err := file.CreateDirs(path); err != nil {
			return err
		}
	}
	return
}
