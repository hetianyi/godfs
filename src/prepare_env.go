package main

import (
	"os"
	"util/file"
)

func main() {
	pwd, _ := os.Getwd()
	file.CopyFile(pwd+"/go_build_storage_go.exe", "E:/godfs-storage/storage1/bin/go_build_storage_go.exe")
	file.CopyFile(pwd+"/go_build_storage_go.exe", "E:/godfs-storage/storage2/bin/go_build_storage_go.exe")

	file.CopyFile(pwd+"/go_build_tracker_go.exe", "E:/godfs-storage/tracker/bin/go_build_tracker_go.exe")
	file.CopyFile(pwd+"/go_build_client_go.exe", "E:/godfs-storage/client/bin/go_build_client_go.exe")
}
