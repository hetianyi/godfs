package lib_client

import (
    "testing"
    "fmt"
    "regexp"
    "util/file"
)

func Test1(t *testing.T) {
    fmt.Println(Upload("D:/UltraISO.zip"))
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