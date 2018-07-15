package main

import (
    "util/file"
    "net"
    "util/logger"
    "time"
    "lib_common/header"
    "lib_common"
    "bytes"
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
                    len, e := fi.Read(buff)
                    if len > 0 {
                        conn.Write(buff[0:len])
                    } else {
                        logger.Error(e)
                        break
                    }
                }
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
    path := "D:/nginx-1.8.1.zip"
    Upload(path)

}
