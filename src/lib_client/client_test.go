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
    "crypto/md5"
    "encoding/hex"
    "time"
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
    //fmt.Println(client.Upload("D:/library.zip"))
    //fmt.Println(client.Upload("D:/FTP/instantfap-gifs.part8.zip"))
    //fmt.Println(client.Upload("D:/图片/图片.rar"))
    fmt.Println("本地计算Md5：" + FileHash("D:/IMG_20161207_155837.jpg"))
    fmt.Println(client.Upload("D:/IMG_20161207_155837.jpg"))
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
    path := "/G01/002/M/0d3cc782c3242cf3ce4b2174e1041ed2"

    /*fmt.Println(client.DownloadFile(path, 0, 1024*1024, func(fileLen uint64, reader io.Reader) error {
        newFile, _ := file.CreateFile("E:/godfs-storage/123.jpg")
        defer newFile.Close()
        d := make([]byte, fileLen)
        io.ReadFull(reader, d)
        newFile.Write(d)
        return nil
    }))
    client.DownloadFile(path, 1024*1024, 1024*1024, func(fileLen uint64, reader io.Reader) error {
        newFile, _ := file.OpenFile4Write("E:/godfs-storage/123.jpg")
        defer newFile.Close()
        d := make([]byte, fileLen)
        io.ReadFull(reader, d)
        newFile.Write(d)
        return nil
    })*/
    client.DownloadFile(path, 0, -1, func(fileLen uint64, reader io.Reader) error {
        newFile, _ := file.OpenFile4Write("E:/godfs-storage/123.jpg")
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
        ll,_ := bridge.ReadBytes(buffer, len, body, nil)
        fmt.Println(ll)
        fi, _ := file.CreateFile("D:/godfs/data/tmp/" + timeutil.GetUUID() + ".rar")
        fi.Write(buffer)
        fi.Close()
        body.Close()

        break
    }

}

func FileHash(path string) string {
    fi, e := file.GetFile(path)
    defer fi.Close()
    if e != nil {
        logger.Error(e)
    } else {
        // calculate md5
        md := md5.New()
        buffer := make([]byte, 1024000)
        for {
            len, e1 := fi.Read(buffer)
            if e1 == nil {
                md.Write(buffer[0:len])
            } else {
                break
            }
        }
        sliceCipherStr := md.Sum([]byte(""))
        md5  := hex.EncodeToString(sliceCipherStr)
        return md5
    }
    return ""
}


func Test6(t *testing.T) {
    fmt.Println(FileHash("D:/IMG_20161207_155837.jpg"))
    //fmt.Println(FileHash("D:/UltraISO.zip"))
    //fmt.Println(FileHash("E:/godfs-storage/storage1/data/B3/39/b339ab019dbe6ef3d1b77784c2aa6732"))
    fmt.Println([]byte{1,2,3,4,5,6}[1:3])
}

func Test7(t *testing.T) {

    out := make(chan int)

    for i := 0; i < 50; i++ {
        go func() {
            start := timeutil.GetMillionSecond(time.Now())
            client := Init()
            path := "/G01/002/M/0d3cc782c3242cf3ce4b2174e1041ed2"
            fname := timeutil.GetUUID()
            fmt.Println(client.DownloadFile(path, 0, 1029, func(fileLen uint64, reader io.Reader) error {
                newFile, _ := file.OpenFile4Write("E:/godfs-storage/test/"+ fname +".jpg")
                defer newFile.Close()
                d := make([]byte, fileLen)
                io.ReadFull(reader, d)
                newFile.Write(d)
                return nil
            }))
            fmt.Println(client.DownloadFile(path, 1029, -1, func(fileLen uint64, reader io.Reader) error {
                newFile, _ := file.OpenFile4Write("E:/godfs-storage/test/"+ fname +".jpg")
                defer newFile.Close()
                d := make([]byte, fileLen)
                io.ReadFull(reader, d)
                newFile.Write(d)
                return nil
            }))
            end := timeutil.GetMillionSecond(time.Now())
            logger.Info("下载完成：" + fname, ", 用时：", end-start, "ms")
            out <- 1
        }()
    }
    for i := 0; i < 50; i++ {
        <-out
    }
}
