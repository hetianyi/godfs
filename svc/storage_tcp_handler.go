package svc

import (
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
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

func StartStorageTcpServer() {
	listener, err := net.Listen("tcp", common.InitializedStorageConfiguration.BindAddress+":"+convert.IntToStr(common.InitializedStorageConfiguration.Port))
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Millisecond * 50)
	logger.Info(" tcp server listening on ", common.InitializedStorageConfiguration.BindAddress, ":", common.InitializedStorageConfiguration.Port)
	logger.Info(aurora.BrightGreen("::: storage server started :::"))
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
				h, b, l, err := authenticationHandler(header, common.InitializedStorageConfiguration.Secret)
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
				h, b, l, err := uploadFileHandler(bodyReader, bodyLength)
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
			}
			return pip.Send(&common.Header{
				Result:     common.SUCCESS,
				Msg:        "",
				Attributes: map[string]string{"Name": "李四"},
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

func authenticationHandler(header *common.Header, secret string) (*common.Header, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, 0, nil
	}
	s := header.Attributes["secret"]
	if s != secret {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, 0, nil
	}
	return &common.Header{
		Result: common.SUCCESS,
		Msg:    "authentication success",
	}, nil, 0, nil
}

func uploadFileHandler(bodyReader io.Reader, bodyLength int64) (*common.Header, io.Reader, int64, error) {
	buffer := make([]byte, common.BUFFER_SIZE)
	var realRead int64 = 0
	crcH := util.CreateCrc32Hash()
	md5H := util.CreateMd5Hash()
	tmpFileName := common.InitializedStorageConfiguration.TmpDir + "/" + uuid.UUID()
	out, err := file.CreateFile(tmpFileName)
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() {
		out.Close()
		file.Delete(tmpFileName)
	}()
	for true {
		n, err := bodyReader.Read(buffer)
		realRead += int64(n)
		if err != nil && err != io.EOF {
			return nil, nil, 0, err
		}
		if n > 0 {
			_, err := crcH.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
			_, err = md5H.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
			_, err = out.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
		} else {
			out.Close()
			if bodyLength != realRead {
				return nil, nil, 0, errors.New("mismatch body length")
			}
			crc32String := util.GetCrc32HashString(crcH)
			md5String := util.GetMd5HashString(md5H)

			targetDir := strings.ToUpper(strings.Join([]string{crc32String[len(crc32String)-4 : len(crc32String)-2], "/",
				crc32String[len(crc32String)-2:]}, ""))
			// 文件放在crc结尾的目录，防止目恶意伪造md5文件进行覆盖
			// 避免了暴露文件md5可能出现的风险：保证了在md5相等但是文件不同情况下文件出现的覆盖情况。
			// 此时要求文件的交流必须携带完整的参数
			targetLoc := common.InitializedStorageConfiguration.DataDir + "/" + targetDir
			targetFile := common.InitializedStorageConfiguration.DataDir + "/" + targetDir + "/" + md5String
			// TODO fileId need to be redesign.
			// md5 + crc end + ts + size + srcnode
			// ts: for download
			// ref: http://blog.chinaunix.net/uid-20196318-id-4058561.html
			// another consideration is that the file may be duplicated。
			finalFileId := common.InitializedStorageConfiguration.Group + "/" + targetDir + "/" + md5String
			if !file.Exists(targetLoc) {
				if err := file.CreateDirs(targetLoc); err != nil {
					return nil, nil, 0, err
				}
			}
			if file.Exists(targetFile) {
				return &common.Header{
					Result: common.SUCCESS,
					Attributes: map[string]string{
						"fid":        finalFileId,
						"instanceId": common.InitializedStorageConfiguration.InstanceId,
						"group":      common.InitializedStorageConfiguration.Group,
					},
				}, nil, 0, nil
			}
			if err := file.MoveFile(tmpFileName, targetFile); err != nil {
				return nil, nil, 0, err
			}
			return &common.Header{
				Result: common.SUCCESS,
				Attributes: map[string]string{
					"fid":        finalFileId,
					"instanceId": common.InitializedStorageConfiguration.InstanceId,
					"group":      common.InitializedStorageConfiguration.Group,
				},
			}, nil, 0, nil
		}
	}

	return &common.Header{}, nil, 0, nil
}

func downFileHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	var offset int64 = 0
	var length int64 = -1
	var fileId = header.Attributes["fileId"]
	// parse fileId
	if !common.FileIdPatternRegexp.Match([]byte(fileId)) || header.Attributes == nil {
		return &common.Header{
			Result: common.NOT_FOUND,
		}, nil, 0, nil
	}
	to, err := convert.StrToInt64(header.Attributes["offset"])
	if err != nil {
		offset = 0
	} else {
		offset = to
	}
	tl, err := convert.StrToInt64(header.Attributes["length"])
	if err != nil {
		length = -1
	} else {
		length = tl
	}

	// group := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$1")
	p1 := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$2")
	p2 := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$3")
	md5 := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$4")
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
	if offset >= info.Size() {
		offset = info.Size()
	}
	if length == -1 || offset+length >= info.Size() {
		length = info.Size() - offset
	}
	return &common.Header{
		Result: common.SUCCESS,
	}, io.LimitReader(fi, length), length, nil
}

// inspectFileHandler inspects file's information
func inspectFileHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	var fileId = header.Attributes["fileId"]
	// parse fileId
	if !common.FileIdPatternRegexp.Match([]byte(fileId)) || header.Attributes == nil {
		return &common.Header{
			Result: common.NOT_FOUND,
		}, nil, 0, nil
	}
	// group := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$1")
	p1 := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$2")
	p2 := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$3")
	md5 := common.FileIdPatternRegexp.ReplaceAllString(fileId, "$4")
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
	finfo := &common.FileInfo{
		Group:      common.InitializedStorageConfiguration.Group,
		FileId:     fileId,
		FileLength: info.Size(),
		CreateTime: gox.GetTimestamp(info.ModTime()),
	}
	bs, _ := json.Marshal(finfo)
	return &common.Header{
		Result:     common.SUCCESS,
		Attributes: map[string]string{"info": string(bs)},
	}, nil, 0, nil
}
