package util

import (
	"fmt"
	"github.com/hetianyi/godfs/common"
)

func PrintLogo() {
	/*fmt.Print(`
	     ___    __  __   _________
	    /   \  /  \ |  \ |    /
	   (   --╮(    )|   )|━━━ ╰-╮
	    \___.╯ \__/ |__/ |   __,╯
	`)*/
	fmt.Print(`
   ______  ______  _____    _____  _____
  / ____/ / __  / / ___ ╮  / ___/ / ___   GoDFS::v` + common.VERSION + `
 / /_/ / / /_/ / / /__/ ) / /__/ /__  /   A distribute filesystem.
/_____/ /_____/ /____.·╯ /_/ ________/    github.com/hetianyi/godfs

`)
}
