package common

import (
	"math/rand"
	"time"
	"strconv"
)

var seeds = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// create simple uuid from rand seed
func UUID() string {
	var buffer = make([]rune, 30)
	for i := 0; i < 30; i++ {
		index := rand.Int31n(30)
		buffer[i] = seeds[index]
	}
	return string(buffer)
}


// encode input string to ascii
func EncodeASCII(input string) string {
	ret := strconv.QuoteToASCII(input)
	return ret[1:len(ret) - 1]
}

// decode result of EncodeASCII()
func DecodeASCII(input string) string {
	ret, _ := strconv.Unquote("\"" + input + "\"")
	return ret
}
