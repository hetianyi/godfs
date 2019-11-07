package util

import (
	"fmt"
	"github.com/hetianyi/godfs/common"
)

func PrintLogo() {
	fmt.Print(`
   ____  ____  ____  _________
  / ___\/ __ \/ __ \/ ___/ __/ GoDFS::v` + common.VERSION + `
 / /_/\  /_/ / /_/ / /__/\ \   A distribute filesystem.
 \____/\____/____./_/______/   github.com/hetianyi/godfs

`)
}
