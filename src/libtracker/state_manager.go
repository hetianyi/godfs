package libtracker

import (
	"app"
	"container/list"
	"libcommon/bridge"
	"strconv"
	"sync"
	"time"
	"util/logger"
	"util/timeutil"
	"regexp"
	"util/common"
)

var managedStorages = make(map[string]*storageMeta)

var operationLock = *new(sync.Mutex)
const ipv4Pattern          = "^((25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)$"
const ipv4WithPortPattern  = "^(((25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[0-1]?\\d\\d?)):([0-9]{1,5})$"

type storageMeta struct {
	ExpireTime int64
	UUID       string
	Group      string
	InstanceId string
	Host       string
	Port       int
	HttpPort   int
	HttpEnable bool
	// 统计信息
	TotalFiles int
	Finish     int
	StartTime  int64
	Downloads  int
	Uploads    int
	IOin       int64
	IOout      int64
	DiskUsage  int64
	Memory     uint64
	ReadOnly   bool
}

// 定时任务，剔除过期的storage服务器
func ExpirationDetection() {
	timer := time.NewTicker(app.STORAGE_CLIENT_EXPIRE_TIME)
	for {
		<-timer.C
		logger.Debug("exec expired detected")
		curTime := time.Now().UnixNano() / 1e6
		operationLock.Lock()
		common.Try(func() {
			for k, v := range managedStorages {
				if v.ExpireTime <= curTime { // 过期
					delete(managedStorages, k)
					logger.Info("storage server:", k, "expired finally")
				}
			}
		}, func(i interface{}) {})
		operationLock.Unlock()
	}
}

// 添加storage服务器
func AddStorageServer(meta *bridge.OperationRegisterStorageClientRequest) {
	operationLock.Lock()
	defer operationLock.Unlock()
	host, port := parseAdvertiseAddr(meta.AdvertiseAddr, meta.Port)
	key := host + ":" + strconv.Itoa(port)
	holdMeta := &storageMeta{
		UUID:		meta.UUID,
		ExpireTime: timeutil.GetTimestamp(time.Now().Add(time.Hour * 876000)), // set to 100 years
		Group:      meta.Group,
		InstanceId: meta.InstanceId,
		Host:       host,
		Port:       port,
		HttpPort:   meta.HttpPort,
		HttpEnable: meta.HttpEnable,
		TotalFiles: meta.TotalFiles,
		Finish:     meta.Finish,
		IOin:       meta.IOin,
		IOout:      meta.IOout,
		DiskUsage:  meta.DiskUsage,
		Downloads:  meta.Downloads,
		Uploads:    meta.Uploads,
		StartTime:  meta.StartTime,
		Memory:     meta.Memory,
		ReadOnly:   meta.ReadOnly,
	}
	if managedStorages[key] == nil {
		logger.Debug("register storage server:", key)
	}
	managedStorages[key] = holdMeta
	//js, _ := json.Marshal(*managedStorages[key])
	//fmt.Println(string(js))
}

// 执行即将过期storage服务器
// 通常是storage客户端和tracker服务器断开连接时
func FutureExpireStorageServer(meta *bridge.OperationRegisterStorageClientRequest) {
	operationLock.Lock()
	defer operationLock.Unlock()
	if meta != nil {
		host, port := parseAdvertiseAddr(meta.AdvertiseAddr, meta.Port)
		key := host + ":" + strconv.Itoa(port)
		logger.Info("expire storage client:", key, "in", app.STORAGE_CLIENT_EXPIRE_TIME)
		holdMeta := &storageMeta{
			UUID:		meta.UUID,
			ExpireTime: timeutil.GetTimestamp(time.Now().Add(app.STORAGE_CLIENT_EXPIRE_TIME)),
			Group:      meta.Group,
			InstanceId: meta.InstanceId,
			Host:       host,
			Port:       port,
			HttpPort:   meta.HttpPort,
			HttpEnable: meta.HttpEnable,
			TotalFiles: meta.TotalFiles,
			Finish:     meta.Finish,
			IOin:       meta.IOin,
			IOout:      meta.IOout,
			DiskUsage:  meta.DiskUsage,
			Downloads:  meta.Downloads,
			Uploads:    meta.Uploads,
			StartTime:  meta.StartTime,
			Memory:     meta.Memory,
			ReadOnly:   meta.ReadOnly,
		}
		managedStorages[key] = holdMeta
	}
}

// check if instance if is unique
func IsInstanceIdUnique(meta *bridge.OperationRegisterStorageClientRequest) bool {
	operationLock.Lock()
	defer operationLock.Unlock()
	host, port := parseAdvertiseAddr(meta.AdvertiseAddr, meta.Port)
	key := host + ":" + strconv.Itoa(port)
	for k, v := range managedStorages {
		if k != key && v.Group == meta.Group && v.InstanceId == meta.InstanceId {
			return false
		}
	}
	return true
}

// 获取组内成员
func GetGroupMembers(meta *bridge.OperationRegisterStorageClientRequest) []bridge.Member {
	operationLock.Lock()
	defer operationLock.Unlock()
	host, port := parseAdvertiseAddr(meta.AdvertiseAddr, meta.Port)
	key := host + ":" + strconv.Itoa(port)
	var mList list.List
	for k, v := range managedStorages {
		if k != key && v.Group == meta.Group { // 过期
			m := bridge.Member{AdvertiseAddr: v.Host, Port: v.Port, InstanceId: v.InstanceId, Group: v.Group, ReadOnly: v.ReadOnly, HttpEnable: v.HttpEnable, HttpPort: v.HttpPort}
			mList.PushBack(m)
		}
	}
	var members = make([]bridge.Member, mList.Len())
	index := 0
	for e := mList.Front(); e != nil; e = e.Next() {
		members[index] = e.Value.(bridge.Member)
		index++
	}
	return members
}

// 获取组内成员
func GetAllStorages() []bridge.Member {
	operationLock.Lock()
	defer operationLock.Unlock()
	var mList list.List
	for _, v := range managedStorages {
		m := bridge.Member{AdvertiseAddr: v.Host, Port: v.Port, InstanceId: v.InstanceId, Group: v.Group, ReadOnly: v.ReadOnly, HttpEnable: v.HttpEnable, HttpPort: v.HttpPort}
		mList.PushBack(m)
	}
	var members = make([]bridge.Member, mList.Len())
	index := 0
	for e := mList.Front(); e != nil; e = e.Next() {
		members[index] = e.Value.(bridge.Member)
		index++
	}
	return members
}

func GetSyncStatistic() []bridge.ServerStatistic {
	operationLock.Lock()
	defer operationLock.Unlock()
	var res = make([]bridge.ServerStatistic, len(managedStorages))
	i := 0
	for _, v := range managedStorages {
		item := bridge.ServerStatistic {
			UUID:		   v.UUID,
			AdvertiseAddr: v.Host,
			Group:         v.Group,
			InstanceId:    v.InstanceId,
			Port:          v.Port,
			HttpPort:      v.HttpPort,
			HttpEnable:    v.HttpEnable,
			TotalFiles:    v.TotalFiles,
			Finish:        v.Finish,
			IOin:          v.IOin,
			IOout:         v.IOout,
			DiskUsage:     v.DiskUsage,
			Downloads:     v.Downloads,
			Uploads:       v.Uploads,
			StartTime:     v.StartTime,
			Memory:        v.Memory,
			ReadOnly:      v.ReadOnly,
		}
		res[i] = item
		i++
	}
	return res
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
