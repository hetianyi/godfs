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
    "app"
    "container/list"
)

func Init() *Client {
    logger.SetLogLevel(1)
    app.TRACKERS = "192.168.0.104:1022"
    app.SECRET = "OASAD834jA97AAQE761=="
    app.GROUP = "G01"
    client:= NewClient(10)
    collector := TaskCollector{
        Interval: time.Second * 30,
        FirstDelay: 0,
        Name: "同步storage server",
        Job: SyncAllStorageServersTaskCollector,
    }
    collectors := *new(list.List)
    collectors.PushBack(&collector)
    maintainer := &TrackerMaintainer{Collectors: collectors}
    maintainer.Maintain(app.TRACKERS)
    return client
}


func Test1(t *testing.T) {
    client := Init()
    //fmt.Println(client.Upload("D:/UltraISO.zip")) // G01/002/M/432597de0e65eedbc867620e744a35ad
    //fmt.Println(client.Upload("F:/project.rar"))
    //fmt.Println(client.Upload("D:/library.zip"))
    //fmt.Println(client.Upload("D:/FTP/instantfap-gifs.part8.zip"))
    //fmt.Println(client.Upload("D:/图片/图片.rar"))
    fmt.Println("本地计算Md5：" + FileHash("D:/UltraISO.zip"))
    time.Sleep(time.Second)
    //fmt.Println(client.Upload("D:/UltraISO.zip", app.GROUP))
    //fmt.Println(client.Upload("D:/1114785.jpg", app.GROUP))
    //fmt.Println(client.Upload("F:/project.rar", app.GROUP))
    //fmt.Println(client.Upload("F:/Software/AtomSetup-1.18.0_x64.exe", app.GROUP))
    fmt.Println(client.Upload("D:/VMWare/ISO/CentOS-7-x86_64-Everything-1708.iso", app.GROUP))
}


func Test2(t *testing.T) {
    client := Init()
    fmt.Println(client.QueryFile("432597de0e65eedbc867620e744a35ad"))
}

func Test3(t *testing.T) {
    regex := "^/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})$"
    value := regexp.MustCompile(regex).ReplaceAllString("/x_/_123/432597de0e65eedbc867620e744a35ad", "${3}")
    fmt.Println(value)
}


func Test4(t *testing.T) {
    client := Init()
    path := "/G01/002/M/432597de0e65eedbc867620e744a35ad"

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
        newFile, _ := file.CreateFile("E:/godfs-storage/123.zip")
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

func Test8(t *testing.T) {
    client := Init()
    path := "/G01/001/S/061af19e7aebbf4a159664a8b96a13cd"
    for i := 0; i < 10 ; i++ {
        go func() {
            for  {
                client.DownloadFile(path, 0, -1, func(fileLen uint64, reader io.Reader) error {
                    newFile, _ := file.CreateFile("C:/Users/Tisnyi/Downloads/123/" + timeutil.GetUUID() + ".jpg")
                    d := make([]byte, fileLen)
                    io.ReadFull(reader, d)
                    newFile.Write(d)
                    newFile.Close()
                    logger.Info("finish")
                    return nil
                })
            }
        }()
    }
    cha := make(chan int)
    <- cha
}



func paramT(a *[]byte) {
    println(a)
}
func paramS(s *S) {
    println(s)
}

type S struct {
    Name string
}

func Test9(t *testing.T) {
    a := make([]byte, 1024)
    a[0] = 2
    a[1] = 3
    a[2] = 4
    a[3] = 5
    a[4] = 6

    s := S{
       Name: "lisi" ,
    }


    println(&a)
    println(&s)
    paramT(&a)
    paramS(&s)
}
