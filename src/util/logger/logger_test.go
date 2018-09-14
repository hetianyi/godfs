package logger

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	fmt.Println()
}

func Test5(t *testing.T) {
	/*const size = 64 << 10
	  buf := make([]byte, size)
	  runtime.StartTrace()
	  buf = buf[:runtime.Stack(buf, false)]
	  fmt.Println(string(buf))*/

	Info("knasdnaksndalsdllkadlasasldn阿里斯顿你懒得拉上来阿里哪里23123123")
}
