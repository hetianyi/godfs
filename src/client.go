package main

import (
    "util/file"
    "net"
    "util/logger"
    "time"
    "lib_common/header"
    "lib_common"
    "bytes"
    "io"
    "encoding/json"
    "flag"
    "fmt"
)

//client demo for upload file to storage server.
func Upload(path string) error {

    conn, e := net.Dial("tcp", "127.0.0.1:1024")
    if e == nil {
        for {
            fi, e := file.GetFile(path)
            if e == nil {
                fInfo, _ := fi.Stat()

                operation := 2
                meta := &header.UploadRequestMeta{
                    Secret: "OASAD834jA97AAQE761==",
                    FileSize: fInfo.Size(),
                }

                metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(fInfo.Size(), meta)
                if e2 != nil {
                    logger.Fatal("meta prepare failed")
                }

                var headerBuff bytes.Buffer
                headerBuff.Write(header.OperationHeadByteMap[operation])
                headerBuff.Write(metaSize)
                headerBuff.Write(bodySize)

                len1, e2 := conn.Write(headerBuff.Bytes())
                if e2 != nil || len1 != headerBuff.Len() {
                    logger.Fatal("error write meta len")
                }
                len2, e3 := conn.Write(metaBytes)
                if e3 != nil || len2 != len(metaBytes) {
                    logger.Fatal("error write meta")
                }

                buff := make([]byte, 1024*30)
                for {
                    len5, e := fi.Read(buff)
                    if len5 > 0 {
                        len3, e4 := conn.Write(buff[0:len5])
                        if e4 != nil || len3 != len(buff[0:len5]) {
                            lib_common.Close(conn)
                            logger.Fatal("error write body:", e4)
                        }
                    } else {
                        if e != io.EOF {
                            lib_common.Close(conn)
                            logger.Error(e)
                        } else {
                            logger.Info("上传完毕")
                        }
                        break
                    }
                }
                _, respMeta, _, e6 := lib_common.ReadConnMeta(conn)
                if e6 != nil {
                    logger.Fatal("error read response:", e6)
                }
                var resp = &header.UploadResponseMeta{}
                e7 := json.Unmarshal([]byte(respMeta), resp)
                if e7 != nil {
                    lib_common.Close(conn)
                    logger.Error(e7)
                }
                logger.Info(respMeta)
            }
            time.Sleep(time.Millisecond * 10)
            //break
        }
    } else {
        logger.Error("error connect to storage server")
    }
    return e
}

func main() {

    var uploadFile = flag.String("f", "", "custom config file")
    flag.Parse()


    fmt.Println("上传文件：", *uploadFile)

   /* path := "D:/nginx-1.8.1.zip"
    Upload(path)*/

}
