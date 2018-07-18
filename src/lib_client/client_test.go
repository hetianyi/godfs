package lib_client

import (
    "testing"
    "fmt"
    "regexp"
    "util/file"
    "net/http"
    "strconv"
    "util/timeutil"
    "lib_common"
)

func Test1(t *testing.T) {
    //fmt.Println(Upload("D:/UltraISO.zip"))
    //fmt.Println(Upload("F:/project.rar"))
    //fmt.Println(Upload("D:/nginx-1.8.1.zip"))
    //fmt.Println(Upload("D:/图片/图片.rar"))
    fmt.Println(Upload("D:/IMG_20161207_155837.jpg"))
}


func Test2(t *testing.T) {
    fmt.Println(CheckFileExists("432597de0e65eedbc867620e744a35ad"))
}

func Test3(t *testing.T) {
    regex := "^/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})$"
    value := regexp.MustCompile(regex).ReplaceAllString("/x_/_123/432597de0e65eedbc867620e744a35ad", "${3}")
    fmt.Println(value)
}


func Test4(t *testing.T) {
    path := "/G1/001/432597de0e65eedbc867620e744a35ac"
    newFile, _ := file.CreateFile("D:/godfs/test_down/432597de0e65eedbc867620e744a35ad.zip")
    fmt.Println(DownloadFile(path, newFile))
}

func Test5(t *testing.T) {
    for {
        conn,_ := http.Get("http://localhost:8001/download/G01/001/a9a79cfdf784946e72a079c317a0a7c9")
        body := conn.Body
        len, _ := strconv.Atoi(conn.Header.Get("Content-Length"))
        var buffer = make([]byte, len)
        ll,_ := lib_common.ReadBytes(buffer, len, body)
        fmt.Println(ll)
        fi, _ := file.CreateFile("D:/godfs/data/tmp/" + timeutil.GetUUID() + ".rar")
        fi.Write(buffer)
        fi.Close()
        body.Close()

        break
    }

}
