package app

import (
	"container/list"
	"strconv"
	"strings"
)

// include all ORM struct

// table files
type FileDO struct {
	Id         int64  `gorm:"column:id;auto_increment;primary_key" json:"id"`
	Md5        string `gorm:"column:md5" json:"md5"`
	PartNumber int    `gorm:"column:parts_num" json:"parts_num"`
	Group      string `gorm:"column:grop" json:"group"`
	Instance   string `gorm:"column:instance" json:"instance"`
	Finish     int    `gorm:"column:finish" json:"finish"`
	FileSize   int64  `gorm:"column:file_size" json:"file_size"`
	Flag       int    `gorm:"column:flag" json:"flag"` // flag, 0:public, 1:private
}

func (FileDO) TableName() string {
	return "file"
}

// table clients
type StorageDO struct {
	Uuid          string `gorm:"primary_key" json:"uuid"`
	Host          string `gorm:"column:host" json:"host"`
	Port          int    `gorm:"column:port" json:"port"`
	AdvertiseAddr string `gorm:"column:advertise_addr" json:"advertise_addr"`
	AdvertisePort int    `gorm:"column:advertise_port" json:"advertise_port"`
	Status        int    `gorm:"column:status" json:"status"`
	Group         string `gorm:"column:grop" json:"group"`
	InstanceId    string `gorm:"column:instance_id" json:"instance_id"`
	HttpPort      int    `gorm:"column:http_port" json:"http_port"`
	HttpEnable    bool   `gorm:"column:http_enable" json:"http_enable"`
	StartTime     int64  `gorm:"column:start_time" json:"start_time"`
	Download      int64  `gorm:"column:downloads" json:"downloads"`
	Upload        int64  `gorm:"column:uploads" json:"uploads"`
	Disk          int64  `gorm:"column:disk" json:"disk"`
	ReadOnly      bool   `gorm:"column:read_only" json:"read_only"`
	TotalFiles    int    `gorm:"column:total_files" json:"total_files"`
	Finish        int    `gorm:"column:finish" json:"finish"`
	IOin          int64  `gorm:"column:ioin" json:"ioin"`
	IOout         int64  `gorm:"column:ioout" json:"ioout"`
	// 1: use LookBackAddress:Port 2: use AdvertiseAddr:AdvertisePort
	AccessFlag int   `gorm:"column:access_flag" json:"access_flag"`
	LogTime    int64 `gorm:"column:log_time" json:"log_time"`

	// not store in db
	ExpireTime     int64  `gorm:"-" json:"expire_time"`
	StageDownloads int    `gorm:"-" json:"stageDownloads"`
	StageUploads   int    `gorm:"-" json:"stageUploads"`
	StageIOin      int64  `gorm:"-" json:"stageIOin"`
	StageIOout     int64  `gorm:"-" json:"stageIOout"`
	Memory         uint64 `gorm:"-" json:"mem"`
	Secret         string `gorm:"-" json:"secret"`
}

func (StorageDO) TableName() string {
	return "storage"
}

// table part
type PartDO struct {
	Id   int64  `gorm:"column:id;auto_increment;primary_key" json:"id"`
	Md5  string `gorm:"column:md5" json:"md5"`
	Size int64  `gorm:"column:size" json:"size"`
}

func (PartDO) TableName() string {
	return "part"
}

// table relation_file_part
type FilePartRelationDO struct {
	Id     int64 `gorm:"column:id;auto_increment;primary_key" json:"id"`
	FileId int64 `gorm:"column:fid" json:"fid"`
	PartId int64 `gorm:"column:pid" json:"pid"`
}

func (FilePartRelationDO) TableName() string {
	return "relation_file_part"
}

// table sys
type SysDO struct {
	Key   string `gorm:"column:key;primary_key" json:"key"`
	Value string `gorm:"column:value" json:"value"`
}

func (SysDO) TableName() string {
	return "sys"
}

// table tracker
type TrackerDO struct {
	Uuid          string `gorm:"column:uuid;primary_key" json:"uuid"`
	TrackerSyncId int64  `gorm:"column:tracker_sync_id" json:"tracker_sync_id"`
	LastRegTime   int64  `gorm:"column:last_reg_time" json:"last_reg_time"`
	LocalPushId   int64  `gorm:"column:local_push_id" json:"local_push_id"`

	Host       string `gorm:"column:host" json:"host"`
	Port       int    `gorm:"column:port" json:"port"`
	Status     int    `gorm:"column:status" json:"status"` // 0: disabled,  1:enabled, 3: deleted
	Secret     string `gorm:"column:secret" json:"secret"`
	TotalFiles int    `gorm:"column:files" json:"files"`
	Remark     string `gorm:"column:remark" json:"remark"`
	AddTime    int64  `gorm:"column:add_time" json:"add_time"`
}

func (TrackerDO) TableName() string {
	return "tracker"
}

// table web_storage_log
type StorageStatisticLogDO struct {
	Id          int64  `gorm:"column:id;auto_increment;primary_key" json:"id"`
	StorageUuid string `gorm:"column:storage" json:"storage"`
	LogTime     int64  `gorm:"column:log_time" json:"log_time"`
	IOin        int64  `gorm:"column:ioin" json:"ioin"`
	IOout       int64  `gorm:"column:ioout" json:"ioout"`
	Disk        int64  `gorm:"column:disk" json:"disk"`
	Memory      int64  `gorm:"column:mem" json:"mem"`
	Download    int64  `gorm:"column:download" json:"download"`
	Upload      int64  `gorm:"column:upload" json:"upload"`
}

func (StorageStatisticLogDO) TableName() string {
	return "storage_statistic_log"
}

type RelationTrackerStorageDO struct {
	TrackerUuid string `gorm:"column:tracker" json:"tracker"`
	StorageUuid string `gorm:"column:storage" json:"storage"`
}

func (RelationTrackerStorageDO) TableName() string {
	return "relation_tracker_storage"
}

// table files
type FileVO struct {
	Id         int64    `gorm:"column:id;auto_increment;primary_key" json:"id"`
	Md5        string   `gorm:"column:md5" json:"md5"`
	PartNumber int      `gorm:"column:parts_num" json:"parts_num"`
	Group      string   `gorm:"column:grop" json:"group"`
	Instance   string   `gorm:"column:instance" json:"instance"`
	Finish     int      `gorm:"column:finish" json:"finish"`
	FileSize   int64    `gorm:"column:file_size" json:"file_size"`
	Flag       int      `gorm:"column:flag" json:"flag"` // flag, 0:public, 1:private
	Parts      []PartDO `gorm:"-" json:"parts"`
}

func (FileVO) TableName() string {
	return "file"
}

func (vo *FileVO) From(fileDO *FileDO) *FileVO {
	if fileDO == nil {
		return nil
	}
	vo.Id = fileDO.Id
	vo.Md5 = fileDO.Md5
	vo.PartNumber = fileDO.PartNumber
	vo.Group = fileDO.Group
	vo.Instance = fileDO.Instance
	vo.Finish = fileDO.Finish
	vo.FileSize = fileDO.FileSize
	vo.Flag = fileDO.Flag
	return vo
}

// set file parts of fileVO,
// list member must be *PartDO
func (vo *FileVO) SetParts(parts *list.List) {
	if parts == nil {
		return
	}
	temp := make([]PartDO, parts.Len())
	index := 0
	for ele := parts.Front(); ele != nil; ele = ele.Next() {
		temp[index] = *ele.Value.(*PartDO)
		index++
	}
	vo.Parts = temp
}

// set file parts of fileVO,
// list member must be *PartDO
func (vo *FileVO) SetPartsFromVO(parts *list.List) {
	if parts == nil {
		return
	}
	temp := make([]PartDO, parts.Len())
	index := 0
	for ele := parts.Front(); ele != nil; ele = ele.Next() {
		pdo := &PartDO{}
		item := *ele.Value.(*PartVO)
		pdo.Id = item.Id
		pdo.Md5 = item.Md5
		pdo.Size = item.Size
		temp[index] = *pdo
		index++
	}
	vo.Parts = temp
}

// table part
type PartVO struct {
	Id     int64  `gorm:"column:id;auto_increment;primary_key" json:"id"`
	FileId int64  `gorm:"column:fid" json:"fid"`
	Md5    string `gorm:"column:md5" json:"md5"`
	Size   int64  `gorm:"column:size" json:"size"`
}

func (PartVO) TableName() string {
	return "part"
}

// result count struct
type Total struct {
	Count int `gorm:"column:count" json:"count"`
}

// result count struct
type Statistic struct {
	FileCount   int   `gorm:"column:files" json:"files"`
	FinishCount int   `gorm:"column:finish" json:"finish"`
	DiskSpace   int64 `gorm:"column:disk" json:"disk"`
}

type DashboardIndexStatistic struct {
	IOin      int64  `gorm:"column:ioin" json:"ioin"`
	IOout     int64  `gorm:"column:ioout" json:"ioout"`
	Downloads int    `gorm:"column:downloads" json:"downloads"`
	Uploads   int    `gorm:"column:uploads" json:"uploads"`
	Trackers  int    `gorm:"column:trackers" json:"trackers"`
	Storages  int    `gorm:"column:storages" json:"storages"`
	Files     int    `gorm:"column:files" json:"files"`
	UpTime    string `gorm:"column:up_time" json:"up_time"`
}

// server info of storage server and tracker server info
type ServerInfo struct {
	Host          string
	Port          int
	Group         string
	InstanceId    string
	AccessFlag    int
	AdvertiseAddr string
	AdvertisePort int
	IsTracker     bool
	Secret        string
}

func (server *ServerInfo) FromStorage(storage *StorageDO) *ServerInfo {
	server.Host = storage.Host
	server.Port = storage.Port
	server.InstanceId = storage.InstanceId
	server.Group = storage.Group
	server.AccessFlag = storage.AccessFlag
	server.AdvertiseAddr = storage.AdvertiseAddr
	server.AdvertisePort = storage.AdvertisePort
	server.IsTracker = false
	server.Secret = storage.Secret
	return server
}

func (server *ServerInfo) FromTracker(host string, port int, secret string) *ServerInfo {
	server.Host = host
	server.Port = port
	server.IsTracker = true
	server.Secret = secret
	return server
}

func (server *ServerInfo) FromConnStr(connStr string) *ServerInfo {
	server.Host = strings.Split(connStr, ":")[0]
	server.Port, _ = strconv.Atoi(strings.Split(connStr, ":")[1])
	server.IsTracker = true
	server.Secret = Secret
	return server
}

func (server *ServerInfo) SwitchAccessFlag() {
	if server.AccessFlag == AccessFlagInitial {
		server.AccessFlag = AccessFlagAdvertise
	} else {
		server.AccessFlag = AccessFlagInitial
	}
}

func (server *ServerInfo) GetHostAndPortByAccessFlag() (host string, port int) {
	if server.IsTracker {
		server.AdvertiseAddr = server.Host
		server.AdvertisePort = server.Port
		return server.Host, server.Port
	}
	if server.AccessFlag == AccessFlagNone {
		// if run as client, always try from advertise ip
		if RunWith == 3 {
			server.AccessFlag = AccessFlagAdvertise
			return server.AdvertiseAddr, server.AdvertisePort
		}
		server.AccessFlag = AccessFlagInitial
		return server.Host, server.Port
	}
	if server.AccessFlag == AccessFlagInitial {
		return server.Host, server.Port
	}
	return server.AdvertiseAddr, server.AdvertisePort
}

type ClientConfig struct {
	Trackers            []string `json:"trackers"`
	LogLevel            string   `json:"log_level"`
	LogRotationInterval string   `json:"log_rotation_interval"`
	Secret              string   `json:"secret"`
}

// nginx configuration
type NginxStructedConfiguration struct {
	AllServers        list.List
	UploadableServers list.List
	Expose            list.List
}
