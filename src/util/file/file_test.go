package file

import (
    "testing"
    "fmt"
    "regexp"
    "bytes"
)

func Test1(t *testing.T) {
    CopyFileTo("D:/图片/DesktopBackground/farcry4/gamersky_02origin_03_201481320288C2.jpg", "E:/godfs-storage/storage1/data")
}



func Test2(t *testing.T) {
    fmt.Println(GetFileMd5("F:/sphinx-0.9.9-win32-id64-full.zip"))
}

func Test3(t *testing.T) {
    a := []byte{1,2,3,4,6,7}
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
    fmt.Println(regexp.Match(ContentDispositionPattern, []byte("Content-Disposition: form-data; name=\"file2\"; filename=\"F:\\Software\\AtomSetup-1.18.0_x64.exe\""))  )
    name := regexp.MustCompile(ContentDispositionPattern).ReplaceAllString("Content-Disposition: form-data; name=\"file2\"", "${1}")
    fmt.Println(name)
}

