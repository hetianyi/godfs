package common

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(UUID())
	}
}
