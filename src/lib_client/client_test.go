package lib_client

import (
    "testing"
    "fmt"
    "regexp"
    "util/file"
    "net/http"
    "strconv"
    "util/timeutil"
    "util/logger"
    "lib_common/bridge"
    "io"
)

func Init() *Client {
    logger.SetLogLevel(1)
   client, e := NewClient("127.0.0.1", 1024, "OASAD834jA97AAQE761==")
   if e != nil {
       logger.Error(e)
   }
   return client
}


func Test1(t *testing.T) {
    client := Init()
    //fmt.Println(client.Upload("D:/UltraISO.zip")) // G01/002/M/c445b10edc599617106ae8472c1446fd
    //fmt.Println(client.Upload("F:/project.rar"))
    fmt.Println(client.Upload("D:/nginx-1.8.1.zip"))
    //fmt.Println(client.Upload("D:/FTP/instantfap-gifs.part8.zip"))
    //fmt.Println(client.Upload("D:/图片/图片.rar"))
    //fmt.Println(client.Upload("D:/IMG_20161207_155837.jpg"))
}


func Test2(t *testing.T) {
    client := Init()
    fmt.Println(client.CheckFileExists("c445b10edc599617106ae8472c1446f1"))
}

func Test3(t *testing.T) {
    regex := "^/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})$"
    value := regexp.MustCompile(regex).ReplaceAllString("/x_/_123/432597de0e65eedbc867620e744a35ad", "${3}")
    fmt.Println(value)
}


func Test4(t *testing.T) {
    client := Init()
    path := "/G01/002/M/d1cd0a197dc6a5a4beb13cf0fe951444"

    fmt.Println(client.DownloadFile(path, 0, 1024*1024, func(fileLen uint64, reader io.Reader) error {
        newFile, _ := file.CreateFile("D:/godfs/123.zip")
        defer newFile.Close()
        d := make([]byte, fileLen)
        io.ReadFull(reader, d)
        newFile.Write(d)
        return nil
    }))
    client.DownloadFile(path, 1024*1024, 238225, func(fileLen uint64, reader io.Reader) error {
        newFile, _ := file.OpenFile4Write("D:/godfs/123.zip")
        defer newFile.Close()
        d := make([]byte, fileLen)
        io.ReadFull(reader, d)
        newFile.Write(d)
        return nil
    })
}

func Test5(t *testing.T) {
    for {
        conn,_ := http.Get("http://localhost:8001/download/G01/001/a9a79cfdf784946e72a079c317a0a7c9")
        body := conn.Body
        len, _ := strconv.Atoi(conn.Header.Get("Content-Length"))
        var buffer = make([]byte, len)
        ll,_ := bridge.ReadBytes(buffer, len, body)
        fmt.Println(ll)
        fi, _ := file.CreateFile("D:/godfs/data/tmp/" + timeutil.GetUUID() + ".rar")
        fi.Write(buffer)
        fi.Close()
        body.Close()

        break
    }

}
