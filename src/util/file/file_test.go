package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"util/common"
)

func Test1(t *testing.T) {
	CopyFileTo("D:/图片/DesktopBackground/farcry4/gamersky_02origin_03_201481320288C2.jpg", "E:/godfs-storage/storage1/data")
}

func Test2(t *testing.T) {
	fmt.Println(GetFileMd5("F:/sphinx-0.9.9-win32-id64-full.zip"))
}

func Test3(t *testing.T) {
	a := []byte{1, 2, 3, 4, 6, 7}
	fmt.Println(a[1:2])
}
func Test4(t *testing.T) {
	a := "^multipart/form-data; boundary=.*$"
	fmt.Println(regexp.Match(a, []byte("multipart/form-data; boundary=asdasdasd")))
}

func Test5(t *testing.T) {
	fi, _ := GetFile("D:/123.txt")
	buffer := make([]byte, 10240)
	fmt.Println(fi.Read(buffer))
}

func Test6(t *testing.T) {
	a := "123456"
	b := "123"
	fmt.Println(bytes.Index([]byte(a), []byte(b)))
	var buff bytes.Buffer
	c := -1
	buff.Write([]byte(a)[0:c])
}

func Test7(t *testing.T) {
	ContentDispositionPattern := "^Content-Disposition: form-data; name=\"([^\"]+)\"$"
	fmt.Println(regexp.Match(ContentDispositionPattern, []byte("Content-Disposition: form-data; name=\"file2\"; filename=\"F:\\Software\\AtomSetup-1.18.0_x64.exe\"")))
	name := regexp.MustCompile(ContentDispositionPattern).ReplaceAllString("Content-Disposition: form-data; name=\"file2\"", "${1}")
	fmt.Println(name)
}

func TestIsAbsPath(t *testing.T) {
	fmt.Println(filepath.IsAbs("D:/asdasd"))
}

func Test8(t *testing.T) {
	dir := "D:/"
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		fmt.Println(path, common.TValue(info.IsDir(), "d", "-"), info.Size())
		return nil
	})
}

func Test9(t *testing.T) {
	dir := "D:/"
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		return common.TOperation(info.IsDir(), func() interface{} {
			return filepath.SkipDir
		}, func() interface{} {
			fmt.Println(path, common.TValue(info.IsDir(), "d", "-"), info.Size())
			return nil
		}).(error)
	})
}

func Test10(t *testing.T) {
	f := "D:\\FTP\\instantfap-gifs.part7.zip"
	dir := f + string(os.PathSeparator) + ".."
	files, _ := ioutil.ReadDir(dir)
	absPath, _ := filepath.Abs(dir)
	for i := range files {
		fmt.Println(absPath + string(os.PathSeparator) + files[i].Name())
	}
}

func Test11(t *testing.T) {
	p := "asdasd\\123"
	fmt.Println(IsAbsPath(p))
	fmt.Println(filepath.Abs(p))
}
