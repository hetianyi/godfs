package file

import (
    "testing"
    "fmt"
)

func Test1(t *testing.T) {
    CopyFileTo("D:/图片/DesktopBackground/farcry4/gamersky_02origin_03_201481320288C2.jpg", "E:/godfs-storage/storage1/data")
}



func Test2(t *testing.T) {
    fmt.Println(GetFileMd5("F:/sphinx-0.9.9-win32-id64-full.zip"))
}

