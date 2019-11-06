package util

import (
	"fmt"
	"github.com/hetianyi/godfs/common"
)

func PrintLogo() {
	fmt.Print(`
   ____  ____  _____  _____ ___
  / ___\/ __ \/ ___ \/ ___/ ___| GoDFS::v` + common.VERSION + `
 / /_/\  /_/ / /__/ / /__/__  |  A distribute filesystem.
 \____/\____/____, /_/_______/   github.com/hetianyi/godfs

`)
}
