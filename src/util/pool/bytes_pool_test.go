package pool

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestBytesPool1(t *testing.T) {
	for {
		a := make([]byte, 10240)
		fmt.Println(len(a))
		//fmt.Println(a)
		//time.Sleep(1)
	}
}
func TestBytesPool2(t *testing.T) {
	pool := NewBytesPool(5)
	rand.Seed(time.Now().UnixNano())
	for {
		rd := rand.Int31n(10)
		a := pool.Apply(int(rd))
		fmt.Println(len(a))
		pool.Recycle(a)
		time.Sleep(time.Millisecond * 1)
		//fmt.Println(a)
		//time.Sleep(1)
	}
}
