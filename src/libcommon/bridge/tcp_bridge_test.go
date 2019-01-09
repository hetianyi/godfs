package bridge

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	fmt.Print("123")
	fmt.Print(string([]byte{10, 13}))
	fmt.Print("456")
	fmt.Print(rune('\n'))
	fmt.Print(rune('\r'))
	fmt.Print(rune('\r'))
	fmt.Println()
	fmt.Println()
	var a byte = 1
	fmt.Println(int(a))
}

