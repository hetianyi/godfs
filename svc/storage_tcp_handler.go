package svc

import (
	"errors"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/uuid"
	json "github.com/json-iterator/go"
	"github.com/logrusorgru/aurora"
	"io"
	"net"
	"strings"
	"time"
)

var tailRefCount = []byte{0, 0, 0, 1}

func StartStorageTcpServer() {

	listener, err := net.Listen("tcp",
		common.InitializedStorageConfiguration.BindAddress+":"+
			convert.IntToStr(common.InitializedStorageConfiguration.Port))
	if err != nil {
		logger.Fatal(err)
	}

	time.Sleep(time.Millisecond * 50)

	logger.Info(" tcp server listening on ",
		common.InitializedStorageConfiguration.BindAddress, ":",
		common.InitializedStorageConfiguration.Port)
	logger.Info(aurora.BrightGreen("::: storage server started :::"))

	// running in cluster mode.
	if common.InitializedStorageConfiguration.ParsedTrackers != nil &&
		len(common.InitializedStorageConfiguration.ParsedTrackers) > 0 {
		servers := make([]*common.Server, len(common.InitializedStorageConfiguration.ParsedTrackers))
		for i, s := range common.InitializedStorageConfiguration.ParsedTrackers {
			servers[i] = &s
		}
		config := &api.Config{
			MaxConnectionsPerServer: MaxConnPerServer,
			SynchronizeOnce:         false,
			TrackerServers:          servers,
		}
		InitializeClientAPI(config)
		for _, s := range servers {
			go binlogPusher(s)
		}
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting new connection: ", err)
			continue
		}
		logger.Debug("accept a new connection")
		go storageClientConnHandler(conn)
	}
}

func storageClientConnHandler(conn net.Conn) {
	pip := &gpip.Pip{
		Conn: conn,
	}
	defer pip.Close()
	authorized := false
	for {
		err := pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
			if _header == nil {
				return errors.New("invalid request: header is empty")
			}
			header := _header.(*common.Header)
			bs, _ := json.Marshal(header)
			logger.Debug("server got message:", string(bs))
			if header.Operation == common.OPERATION_CONNECT {
				h, _, b, l, err := authenticationHandler(header, common.InitializedStorageConfiguration.Secret)
				if err != nil {
					return err
				}
				if h.Result != common.SUCCESS {
					pip.Send(h, b, l)
					return errors.New("unauthorized connection, force disconnection by server")
				} else {
					authorized = true
					return pip.Send(h, b, l)
				}
			}
			if !authorized {
				pip.Send(&common.Header{
					Result: common.UNAUTHORIZED,
					Msg:    "authentication failed",
				}, nil, 0)
				return errors.New("unauthorized connection, force disconnection by server")
			}
			if header.Operation == common.OPERATION_UPLOAD {
				h, b, l, err := uploadFileHandler(header, bodyReader, bodyLength)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_DOWNLOAD {
				h, b, l, err := downFileHandler(header)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_QUERY {
				h, b, l, err := inspectFileHandler(header)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_SYNC_BINLOGS {
				h, b, l, err := syncBinlogHandler(header)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			}
			return pip.Send(&common.Header{
				Result: common.UNKNOWN_OPERATION,
				Msg:    "unknown operation",
			}, nil, 0)
		})
		if err != nil {
			// shutdown connection error is now disabled
			/*if err != io.EOF {
				logger.Error(err)
			}*/
			pip.Close()
			break
		}
	}
}

func uploadFileHandler(header *common.Header, bodyReader io.Reader, bodyLength int64) (*common.Header, io.Reader, int64, error) {

	logger.Debug("receive file")

	tmpFileName := common.InitializedStorageConfiguration.TmpDir + "/" + uuid.UUID()
	out, err := file.CreateFile(tmpFileName)
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() {
		out.Close()
		file.Delete(tmpFileName)
	}()

	proxy := &DigestProxyWriter{
		crcH: util.CreateCrc32Hash(),
		md5H: util.CreateMd5Hash(),
		out:  out,
	}

	isPrivate := true
	if header.Attributes != nil {
		if header.Attributes["isPrivate"] == "0" {
			isPrivate = false
		}
	}

	logger.Debug("copy file")
	_, err = io.Copy(proxy, io.LimitReader(bodyReader, bodyLength))
	if err != nil {
		return nil, nil, 0, err
	}

	logger.Debug("write tail")
	// write reference count mark.
	_, err = out.Write(tailRefCount)
	if err != nil {
		return nil, nil, 0, err
	}
	out.Close()

	crc32String := util.GetCrc32HashString(proxy.crcH)
	md5String := util.GetMd5HashString(proxy.md5H)

	targetDir := strings.ToUpper(strings.Join([]string{crc32String[len(crc32String)-4 : len(crc32String)-2], "/",
		crc32String[len(crc32String)-2:]}, ""))
	targetLoc := common.InitializedStorageConfiguration.DataDir + "/" + targetDir
	targetFile := common.InitializedStorageConfiguration.DataDir + "/" + targetDir + "/" + md5String
	_finalFileId := common.InitializedStorageConfiguration.Group + "/" + targetDir + "/" + md5String

	logger.Debug("create alias")
	now := time.Now()
	finalFileId := util.CreateAlias(_finalFileId, common.InitializedStorageConfiguration.InstanceId, isPrivate, now)
	if !file.Exists(targetLoc) {
		if err := file.CreateDirs(targetLoc); err != nil {
			return nil, nil, 0, err
		}
	}

	if !file.Exists(targetFile) {
		logger.Debug("file not exists, move to target dir.")
		if err := file.MoveFile(tmpFileName, targetFile); err != nil {
			return nil, nil, 0, err
		}
	} else {
		logger.Debug("file already exists, increasing reference count.")
		// increase file reference count.
		if err = updateFileReferenceCount(targetFile, 1); err != nil {
			return nil, nil, 0, err
		}
	}

	// write binlog.
	logger.Debug("write binlog...")
	if err = writableBinlogManager.Write(binlog.CreateLocalBinlog(finalFileId,
		bodyLength, common.InitializedStorageConfiguration.InstanceId)); err != nil {
		return nil, nil, 0, errors.New("error writing binlog: " + err.Error())
	}

	logger.Debug("add dataset...")
	if err := Add(finalFileId); err != nil {
		return nil, nil, 0, errors.New("error writing dataset: " + err.Error())
	}
	logger.Debug("add dataset success")

	logger.Debug("upload success")

	return &common.Header{
		Result: common.SUCCESS,
		Attributes: map[string]string{
			"fid":      finalFileId,
			"group":    common.InitializedStorageConfiguration.Group,
			"instance": common.InitializedStorageConfiguration.InstanceId,
		},
	}, nil, 0, nil
}

func downFileHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	var offset int64 = 0
	var length int64 = -1
	// TODO duplicate code
	if header.Attributes == nil {
		return &common.Header{
			Result: common.NOT_FOUND,
		}, nil, 0, nil
	}

	to, err := convert.StrToInt64(header.Attributes["offset"])
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	offset = to

	tl, err := convert.StrToInt64(header.Attributes["length"])
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	length = tl

	// parse fileId
	var fileId = header.Attributes["fileId"]
	fileInfo, _, err := util.ParseAlias(fileId, "")
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	fileMeta := fileInfo.Group + "/" + fileInfo.Path
	// group := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$1")
	p1 := common.FileMetaPatternRegexp.ReplaceAllString(fileMeta, "$2")
	p2 := common.FileMetaPatternRegexp.ReplaceAllString(fileMeta, "$3")
	md5 := common.FileMetaPatternRegexp.ReplaceAllString(fileMeta, "$4")
	fullPath := strings.Join([]string{common.InitializedStorageConfiguration.DataDir, p1, p2, md5}, "/")

	readyReader, realLen, err := seekRead(fullPath, offset, length)
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	return &common.Header{
		Result: common.SUCCESS,
	}, readyReader, realLen, nil
}

// inspectFileHandler inspects file's information
func inspectFileHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	// TODO duplicate code
	if header.Attributes == nil {
		return &common.Header{
			Result: common.NOT_FOUND,
		}, nil, 0, nil
	}

	// parse fileId
	var fileId = header.Attributes["fileId"]
	fileInfo, _, err := util.ParseAlias(fileId, "")
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	fileMeta := fileInfo.Group + "/" + fileInfo.Path
	// group := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$1")
	p1 := common.FileMetaPatternRegexp.ReplaceAllString(fileMeta, "$2")
	p2 := common.FileMetaPatternRegexp.ReplaceAllString(fileMeta, "$3")
	md5 := common.FileMetaPatternRegexp.ReplaceAllString(fileMeta, "$4")
	fullPath := strings.Join([]string{common.InitializedStorageConfiguration.DataDir, p1, p2, md5}, "/")
	if !file.Exists(fullPath) {
		return &common.Header{
			Result: common.NOT_FOUND,
		}, nil, 0, nil
	}
	fi, err := file.GetFile(fullPath)
	if !file.Exists(fullPath) {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	info, err := fi.Stat()
	if !file.Exists(fullPath) {
		return &common.Header{
			Result: common.ERROR,
		}, nil, 0, err
	}
	fileInfo.FileLength = info.Size()
	bs, _ := json.Marshal(fileInfo)
	return &common.Header{
		Result:     common.SUCCESS,
		Attributes: map[string]string{"info": string(bs)},
	}, nil, 0, nil
}

// syncBinlogHandler gets local binlogs for other storage server.
func syncBinlogHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    "invalid header(0)",
		}, nil, 0, nil
	}

	// parse fileId
	clientState := header.Attributes["clientState"]
	if clientState == "" {
		return &common.Header{
			Result: common.ERROR,
			Msg:    "invalid header(1)",
		}, nil, 0, nil
	}

	bq := &common.BinlogQueryDTO{}
	if err := json.UnmarshalFromString(clientState, bq); err != nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    "invalid header(2)",
		}, nil, 0, nil
	}
	if bq.FileIndex < 0 || bq.Offset < 0 {
		return &common.Header{
			Result: common.ERROR,
			Msg:    "invalid header(3)",
		}, nil, 0, nil
	}

	// fetch 30+ once a tiem will exceed pip header size
	bls, nOffset, err := writableBinlogManager.Read(bq.FileIndex, bq.Offset, 30)
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    "error query binlog: " + err.Error(),
		}, nil, 0, nil
	}

	result := &common.BinlogQueryResultDTO{}
	result.Logs = bls
	result.FileIndex = bq.FileIndex
	result.Offset = nOffset

	if bq.FileIndex == writableBinlogManager.GetCurrentIndex() && bq.Offset == nOffset {
		result.Offset = bq.Offset
		result.FileIndex = bq.FileIndex
	}

	if writableBinlogManager.GetCurrentIndex() > bq.FileIndex &&
		(bls == nil || len(bls) == 0) {
		result.FileIndex = bq.FileIndex + 1
		result.Offset = 0
	}

	jr, err := json.MarshalToString(result)
	if err != nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    err.Error(),
		}, nil, 0, nil
	}

	return &common.Header{
		Result: common.SUCCESS,
		Attributes: map[string]string{
			"result": jr,
		},
	}, nil, 0, nil
}
