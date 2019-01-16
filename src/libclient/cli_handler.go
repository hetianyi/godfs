package libclient

import (
	"app"
	"container/list"
	"errors"
	"fmt"
	"io"
	"libcommon"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"util/file"
	"util/logger"
	"util/timeutil"
)

// upload files
// paths: file path to be upload
// group: file upload group, if not set, use random group
// skipCheck: whether check md5 before upload
func upload(client *Client, files *list.List, group string, skipCheck bool) error {
	for ele := files.Front(); ele != nil; ele = ele.Next() {
		var startTime = time.Now()
		fid, e := client.Upload(ele.Value.(string), group, startTime, skipCheck)
		if e != nil {
			logger.Error(e)
		} else {
			now := time.Now()
			fmt.Println("[==========] 100% [" + timeutil.GetHumanReadableDuration(startTime, now) + "]\nupload success, file id:")
			fmt.Println("+-------------------------------------------+")
			fmt.Println(fid)
			fmt.Println("+-------------------------------------------+")
		}
	}
	return nil
}

func download(client *Client, path string, customDownloadFileName string) error {
	filePath := ""
	var startTime time.Time
	e := client.DownloadFile(path, 0, -1, func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame, resMeta *bridgev2.DownloadFileResponseMeta) (b bool, e error) {
		path = strings.TrimSpace(path)
		if strings.Index(path, "/") != 0 {
			path = "/" + path
		}
		var fi *os.File
		if customDownloadFileName == "" {
			md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${4}")
			customDownloadFileName = md5
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return true, e1
			}
			fi = f
		} else {
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return true, e1
			}
			fi = f
		}
		defer fi.Close()
		filePath, _ = filepath.Abs(fi.Name())
		startTime = time.Now()
		return true, writeOut(manager.Conn, frame.BodyLength, fi, startTime)
	})
	/*e := client.DownloadFile(path, 0, -1, func(realPath string, fileLen uint64, reader io.Reader) error {
		var fi *os.File
		if customDownloadFileName == "" {
			md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(realPath, "${4}")
			customDownloadFileName = md5
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return e1
			}
			fi = f
		} else {
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return e1
			}
			fi = f
		}
		defer fi.Close()
		filePath, _ = filepath.Abs(fi.Name())
		startTime = time.Now()
		return writeOut(reader, int64(fileLen), fi, startTime)
	})*/
	if e != nil {
		logger.Error("download failed:", e)
		return e
	} else {
		now := time.Now()
		fmt.Println("[==========] 100% [" + timeutil.GetHumanReadableDuration(startTime, now) + "]\ndownload success, file save as:")
		fmt.Println("+-------------------------------------------+")
		fmt.Println(filePath)
		fmt.Println("+-------------------------------------------+")
	}
	return nil
}


func writeOut(in io.Reader, offset int64, out io.Writer, startTime time.Time) error {
	buffer, _ := bridge.MakeBytes(app.BUFF_SIZE, false, 0, false)
	defer bridge.RecycleBytes(buffer)
	var finish, total int64
	var stopFlag = false
	defer func() { stopFlag = true }()
	total = offset
	finish = 0
	go libcommon.ShowPercent(&total, &finish, &stopFlag, startTime)

	// total read bytes
	var readBodySize int64 = 0
	// next time bytes to read
	var nextReadSize int
	for {
		// left bytes is more than a buffer
		if (offset-readBodySize)/int64(len(buffer)) >= 1 {
			nextReadSize = len(buffer)
		} else { // left bytes less than a buffer
			nextReadSize = int(offset - readBodySize)
		}
		if nextReadSize == 0 {
			break
		}
		len, e2 := in.Read(buffer[0:nextReadSize])
		if e2 == nil {
			wl, e5 := out.Write(buffer[0:len])
			if e5 != nil || wl != len {
				return errors.New("error write out")
			}
			finish += int64(len)
			readBodySize += int64(len)
			logger.Trace("write:", readBodySize)
		} else {
			if e2 == io.EOF {
				return nil
			}
			return e2
		}
	}
	return nil
}

