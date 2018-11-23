package common

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
)

const ipv4Pattern = "^((25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)$"
const ipv4WithPortPattern = "^(((25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)):([0-9]{1,5})$"

func Test1(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(UUID())
	}
}

// advAddr: storage configuration parameter 'advertise_addr'
// port: storage real serve port
// return parsed ip address and port
func parseAdvertiseAddr(advAddr string, port int) (string, int) {
	m, e := regexp.Match(ipv4Pattern, []byte(advAddr))
	// if parse error, use serve port and parsed ip address
	if e != nil {
		return "", port
	}
	if m {
		return advAddr, port
	}

	m, e1 := regexp.Match(ipv4WithPortPattern, []byte(advAddr))
	// if parse error, use serve port and parsed ip address
	if e1 != nil {
		return "", port
	}
	if m {
		// 1 5
		regxp := regexp.MustCompile(ipv4WithPortPattern)
		adAddr := regxp.ReplaceAllString(advAddr, "${1}")
		adPort, _ := strconv.Atoi(regxp.ReplaceAllString(advAddr, "${5}"))
		return adAddr, adPort
	}
	return "", port
}

func Test2(t *testing.T) {
	ipv4 := "^((25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)$"
	ipv4WithPort := "^((25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?):([0-9]{1,5})$"

	fmt.Println(regexp.Match(ipv4, []byte("254.168.1.255")))
	fmt.Println(regexp.Match(ipv4WithPort, []byte("254.168.1.255:1233")))
	//fmt.Println(regexp.Match(ipv6, []byte("fe80::407b:bb4a:5980:e35d%15:")))

}

func Test3(t *testing.T) {
	fmt.Println(parseAdvertiseAddr("192.168.0.555", 1234))
	fmt.Println(parseAdvertiseAddr("192.168.0.122:8888", 1234))
}
