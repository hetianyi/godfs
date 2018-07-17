package main

import (
    "flag"
    "fmt"
    "lib_client"
)


func main() {
    var uploadFile = flag.String("f", "", "the file to be uploaded")
    flag.Parse()
    if *uploadFile != "" {
        fmt.Println("上传文件：", *uploadFile)
        lib_client.Upload(*uploadFile)
    }
}
