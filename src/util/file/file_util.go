package file

import (
	"app"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"
	"util/common"
)

func GetFileMd5(fi string) (string, error) {
	md := md5.New()
	f, e := GetFile(fi)
	if e != nil {
		return "", e
	} else {
		defer f.Close()
		_, e1 := io.Copy(md, f)
		if e1 != nil {
			return "", e1
		}
		md5 := hex.EncodeToString(md.Sum(nil))
		return md5, nil
	}
}

// GetFile return a File for read.
func GetFile(path string) (*os.File, error) {
	return os.Open(path)
}

// Copy file.
// 将文件复制为另一个文件
func CopyFile(src string, dest string) (s bool, e error) {
	srcfile, err1 := os.Open(src)
	if err1 == nil {
		// create or truncate dest file
		destfile, err2 := os.OpenFile(dest, syscall.O_CREAT|os.O_WRONLY|syscall.O_TRUNC, 0660)
		// ensure close files finally
		defer func() {
			//log.Print("close src and dest files")
			srcfile.Close()
			destfile.Close()
		}()
		// if "create or truncate dest file" succeed then start copying
		if err2 == nil {
			common.Try(func() {
				bs := make([]byte, app.BUFF_SIZE)
				for {
					len, e1 := srcfile.Read(bs)
					if e1 == io.EOF {
						break
					}
					destfile.Write(bs[0:len])
				}
				s = true
			}, func(i interface{}) {
				s = false
				e = errors.New(fmt.Sprint(i))
			})
		} else {
			s = false
			e = errors.New(fmt.Sprint(err2))
		}
	} else {
		s = false
		e = errors.New("open source file failed")
	}

	return s, e
}

// Copy file to dir.
// 将文件复制到目标文件夹
func CopyFileTo(src string, dir string) (s bool, e error) {
	srcfile, err1 := os.Open(src)
	if err1 == nil {
		// create or truncate dest file
		fileInfo, _ := srcfile.Stat()
		destfile, err2 := os.OpenFile(FixPath(dir)+string(os.PathSeparator)+fileInfo.Name(), syscall.O_CREAT|os.O_WRONLY|syscall.O_TRUNC, 0660)
		// ensure close files finally
		defer func() {
			srcfile.Close()
			destfile.Close()
		}()
		// if "create or truncate dest file" succeed then start copying
		if err2 == nil {
			len, e1 := io.Copy(destfile, srcfile)
			if len != fileInfo.Size() || e1 != nil {
				s = false
				e = e1
			} else {
				s = true
			}
		} else {
			s = false
			e = errors.New(fmt.Sprint(err2))
		}
	} else {
		s = false
		e = errors.New("open source file failed")
	}

	return s, e
}

// check whether the file exists.
// 判断文件是否存在
func Exists(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if fi == nil {
		return false
	}
	return true
}

// remove file or directory.
// if delete failed, you should find out by yourself where and what is the problem.
// special: if the file is not exists, this will return true.
// 删除文件或者文件夹，当删除失败时，需要你自己查出为什么删除失败，
// 特别地：当文件不存在时，返回true
func Delete(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return true
	}
	err = os.Remove(path)
	return nil == err
}

// remove file or directory.
// if delete failed, you should find out by yourself where and what is the problem.
// if is a directory, it will try to delete all files below.
// special: if the file is not exists, this will return true.
// 删除文件或者文件夹，当删除失败时，需要你自己查出为什么删除失败，
// 如果是文件夹则尝试递归删除该文件夹下的所有文件，直到删除失败，
// 特别地：当文件不存在时，返回true
func DeleteAll(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return true
	}
	err = os.RemoveAll(path)
	return nil == err
}

// create new file.
// 创建一个新文件
func CreateFile(path string) (*os.File, error) {
	fi, err := os.Create(path)
	return fi, err
}

// create new file.
// 创建一个新文件
func OpenFile4Write(path string) (*os.File, error) {
	fi, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	return fi, err
}

// create new directory.
// 创建一个新文件夹
func CreateDir(path string) error {
	return os.Mkdir(path, 0777)
}

// create new directory.
// 创建一个新文件夹，如果父文件夹不存在则自动创建
func CreateAllDir(path string) error {
	return os.MkdirAll(path, 0777)
}

// check whether path point to a file.
// 检查路径是否是一个文件
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if nil == err {
		return !fi.IsDir()
	} else {
		return false
	}
}

// check whether path point to a file.
// 检查路径是否是一个文件
func IsFile1(file *os.File) bool {
	fi, err := file.Stat()
	if nil == err {
		return !fi.IsDir()
	} else {
		return false
	}
}

// check whether path point to a file.
// 检查路径是否是一个文件夹
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if nil == err {
		return fi.IsDir()
	} else {
		return false
	}
}

// check whether path point to a file.
// 检查路径是否是一个文件夹
func IsDir1(file *os.File) bool {
	fi, err := file.Stat()
	if nil == err {
		return fi.IsDir()
	} else {
		return false
	}
}

// rename file / move file
// 文件重命名 / 移动文件
func MoveFile(src string, dest string) error {
	return os.Rename(src, dest)
}

// change work directory to path.
// 将当前工作目录修改为dir指定的目录。如果出错，会返回*PathError底层类型的错误
func ChangeWorkDir(path string) error {
	return os.Chdir(path)
}

// get work directory.
// 返回一个对应当前工作目录的根路径。如果当前目录可以经过多条路径抵达（因为硬链接），Getwd会返回其中一个
func GetWorkDir() (string, error) {
	return os.Getwd()
}

// return a temp directory.
// 返回一个用于保管临时文件的默认目录。
func GetTempDir() string {
	return os.TempDir()
}

// 判断给定的路径是否为绝对路径
func IsAbsPath(filePath string) bool {
	return path.IsAbs(filePath)
}

// get file extension
// 获取文件的扩展名
func GetFileExt(filePath string) string {
	return path.Ext(filePath)
}

//fix path, ep:
//    /aaa/aa\\bb\\cc/d/////     -> /aaa/aa/bb/cc/d
//    E:/aaa/aa\\bb\\cc/d////e/  -> E:/aaa/aa/bb/cc/d/e
//    ""                         -> .
//    /                          -> /
func FixPath(input string) string {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return "."
	}
	// replace windows path separator '\' to '/'
	replaceMent := strings.Replace(input, "\\", "/", -1)

	for {
		if strings.Contains(replaceMent, "//") {
			replaceMent = strings.Replace(replaceMent, "//", "/", -1)
			continue
		}
		if replaceMent == "/" {
			return replaceMent
		}
		len := len(replaceMent)
		if len <= 0 {
			break
		}
		if replaceMent[len-1:] == "/" {
			replaceMent = replaceMent[0 : len-1]
		} else {
			break
		}
	}
	return replaceMent
}

// read properties file on filesystem.
func ReadPropFile(path string) (map[string]string, error) {
	f, e := os.Open(path)
	if e == nil {
		if IsFile1(f) {
			propMap := make(map[string]string)
			reader := bufio.NewReader(f)
			for {
				line, e1 := reader.ReadString('\n')
				if e1 == nil || e1 == io.EOF {
					line = strings.TrimSpace(line)
					if len(line) != 0 && line[0] != '#' {
						//li := strings.Split(line, "=")
						eIndex := strings.Index(line, "=")
						if eIndex == -1 {
							return nil, errors.New("error parameter: '" + line + "'")
						}
						li := []string{line[0:eIndex], line[eIndex+1:]}
						if len(li) > 1 {
							k := strings.TrimSpace(li[0])
							v := strings.TrimSpace(joinLeft(li[1:]))
							propMap[k] = v
						} else {
							return nil, errors.New("error parameter: '" + li[0] + "'")
						}
					}
					if e1 == io.EOF {
						break
					}
				} else {
					// real read error.
					return nil, errors.New("error read from configuration file")
				}
			}
			return propMap, nil
		} else {
			return nil, errors.New("expect file path not directory path")
		}
	} else {
		return nil, e
	}
}

// only for ReadPropFile()
func joinLeft(g []string) string {
	if g == nil || len(g) == 0 {
		return ""
	}
	var bf bytes.Buffer
	for i := range g {
		c := strings.Index(g[i], "#")
		if c == -1 {
			bf.WriteString(g[i])
		} else {
			bf.WriteString(g[i][0:c])
			break
		}
	}
	return string(bf.Bytes())
}
