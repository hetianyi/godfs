package common

import (
	"app"
	"container/list"
	"net"
	"strconv"
	"strings"
)



type IPInfo struct {
	Addr string
	InterfaceName string
}

func ConvertBoolFromInt(input int) bool {
	if input <= 0 {
		return false
	}
	return true
}

// convert list to array
func List2Array(ls *list.List) []interface{} {
	if ls == nil {
		return nil
	}
	arr := make([]interface{}, ls.Len())
	index := 0
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		arr[index] = ele.Value
		index++
	}
	return arr
}

// parse host and port from connection string
func ParseHostPortFromConnStr(connStr string) (string, int) {
	host := strings.Split(connStr, ":")[0]
	port, _ := strconv.Atoi(strings.Split(connStr, ":")[1])
	return host, port
}

// ternary operation
func TOperation(condition bool, trueOperation func() interface{}, falseOperation func() interface{}) interface{}{
	if condition {
		if trueOperation == nil {
			return nil
		}
		return trueOperation()
	}
	if falseOperation == nil {
		return nil
	}
	return falseOperation()
}

// ternary operation
func TValue(condition bool, trueValue interface{}, falseValue interface{}) interface{}{
	if condition {
		return trueValue
	}
	return falseValue
}

// walk a list
// walker return value as break signal
// if it is true, break walking
func WalkList(ls *list.List, walker func(item interface{}) bool) {
	if ls == nil {
		return
	}
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		breakWalk := walker(ele.Value)
		if breakWalk {
			break
		}
	}
}

// get self ip address by preferred interface
func GetPreferredIPAddress() string {
	addrs, _ := net.Interfaces()
	var ret list.List
	for i := range addrs {
		info := scan(&addrs[i])
		if info == nil {
			continue
		}
		ret.PushBack(info)
	}
	if app.PREFERRED_NETWORKS.Len() == 0 { // no preferred
		if ret.Len() == 0 {
			return "127.0.0.1"
		}
		// check preferred ip prefix
		if app.PREFERRED_IP_PREFIX != "" {
			matchResult := ""
			WalkList(&ret, func(item interface{}) bool {
				if strings.HasPrefix(item.(*IPInfo).Addr, app.PREFERRED_IP_PREFIX) {
					matchResult = item.(*IPInfo).Addr
					return true
				}
				return false
			})
			if matchResult == "" {
				matchResult = ret.Front().Value.(*IPInfo).Addr
			}
			return matchResult
		}
		return ret.Front().Value.(*IPInfo).Addr
	} else {
		matchResult := ""
		WalkList(&app.PREFERRED_NETWORKS, func(item interface{}) bool {
			iname := item.(string)
			WalkList(&ret, func(item1 interface{}) bool {
				ipInfo := item1.(*IPInfo)
				if ipInfo.InterfaceName == iname {
					matchResult = ipInfo.Addr
					return true
				}
				return false
			})
			if matchResult != "" {
				return true
			}
			return false
		})
		// no match from all existing interfaces
		if matchResult == "" {
			matchResult = ret.Front().Value.(*IPInfo).Addr
			// check preferred ip prefix
			if app.PREFERRED_IP_PREFIX != "" {
				WalkList(&ret, func(item interface{}) bool {
					if strings.HasPrefix(item.(*IPInfo).Addr, app.PREFERRED_IP_PREFIX) {
						matchResult = item.(*IPInfo).Addr
						return true
					}
					return false
				})
			}
		}
		return matchResult
	}
}

func scan(iface *net.Interface) *IPInfo {
	var (
		addr  *net.IPNet
		addrs []net.Addr
		err   error
	)

	if addrs, err = iface.Addrs(); err != nil {
		return nil
	}

	if !strings.Contains(iface.Flags.String(), "up") {
		return nil
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
		return nil
	}

	if addr.IP[0] == 127 {
		return nil
	}

	if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
		return nil
	}

	return &IPInfo{
		Addr: addr.IP.String(),
		InterfaceName: iface.Name,
	}
}
