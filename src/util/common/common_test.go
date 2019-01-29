package common

import (
	"app"
	"container/list"
	"errors"
	"fmt"
	"net"
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

func Test22(t *testing.T) {
	a := TValue(false, "answer is true", "answer is false").(string)
	fmt.Println(a)
}
func Test23(t *testing.T) {
	err := TOperation(false, func() interface{} {
		return errors.New("true error")
	}, func() interface{} {
		return errors.New("false error")
	})
	fmt.Println(err)
}

func TestWalkList(t *testing.T) {
	ls := list.List{}
	ls.PushBack(1)
	ls.PushBack(2)
	ls.PushBack(3)
	WalkList(&ls, func(item interface{}) bool {
		fmt.Println(item)
		return false
	})
}

func TestNet1(t *testing.T) {
	addrs, _ := net.InterfaceAddrs()
	for i := range addrs {
		fmt.Println(addrs[i].String(), "\t\t\t\t", addrs[i].Network())
	}
}
func TestNet2(t *testing.T) {
	addrs, _ := net.Interfaces()
	for i := range addrs {
		scan1(&addrs[i])
	}
}

func scan1(iface *net.Interface) error {
	var (
		addr  *net.IPNet
		addrs []net.Addr
		err   error
	)

	if addrs, err = iface.Addrs(); err != nil {
		return err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				addr = &net.IPNet{
					IP:   ip4,
					Mask: ipnet.Mask[len(ipnet.Mask)-4:],
				}
				break
			}
		}
	}

	if addr == nil {
		return fmt.Errorf("there's no IP network found")
	}

	if addr.IP[0] == 127 {
		return fmt.Errorf("skipping localhost")
	}

	if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
		return fmt.Errorf("mask means network is too large")
	}

	fmt.Println(
		addr.IP.String(),
		iface.Name,
		addr,
	)

	fmt.Println(addr.Mask.Size())

	return nil
}

func TestNet3(t *testing.T) {
	app.PreferredNetworks.PushBack("VMware Network Adapter VMnet1")
	app.PreferredIPPrefix = "192.168.0"
	fmt.Println(GetPreferredIPAddress())
}
