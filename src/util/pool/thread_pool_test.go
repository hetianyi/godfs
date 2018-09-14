package pool

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	var p, _ = NewPool(10, 10)
	i := 0
	p.Exec(func() {
		fmt.Println(i)
	})
}
