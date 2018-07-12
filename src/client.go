package main

import (
    "util/file"
    "common/header"
    "net"
    "util/logger"
    "encoding/json"
    "encoding/binary"
)

//client demo for upload file to storage server.
func Upload(path string) error {

    fi, e := file.GetFile(path)
    if e == nil {
        fInfo, _ := fi.Stat()

        operation := []byte{2,1}
        meta := &header.UploadRequestMeta{
            FileSize: fInfo.Size(),
        }
        metaStr, _ := json.Marshal(meta)
        metaSize := uint64(len([]byte(metaStr)))
        bodySize := uint64(fInfo.Size())

        metaSizeHeader := make([]byte, 8)
        bodySizeHeader := make([]byte, 8)

        binary.BigEndian.PutUint64(metaSizeHeader, metaSize)
        binary.BigEndian.PutUint64(bodySizeHeader, bodySize)

        conn, e := net.Dial("tcp", "127.0.0.1:1024")
        if e == nil {
            conn.Write(operation)
            conn.Write(metaSizeHeader)
            conn.Write(bodySizeHeader)
            conn.Write([]byte(metaStr))

            buff := make([]byte, 1024*30)
            for  {
                len, e := fi.Read(buff)
                if len > 0 {
                    conn.Write(buff[0:len])
                } else {
                    logger.Error(e)
                    break
                }
            }

        } else {
            logger.Error("error connect to storage server")
        }
    }
    return e
}

func main() {
    path := "D:/1.txt"
    Upload(path)
}
