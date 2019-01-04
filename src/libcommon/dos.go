package libcommon

import (
	"time"
)

// include all ORM struct

// table files
type FileDO struct {
	Id int `gorm:"column:id;auto_increment;primary_key"`
	Md5 string `gorm:"column:md5"`
	PartNumber int `gorm:"column:parts_num"`
	Group string `gorm:"column:grop"`
	Instance string `gorm:"column:instance"`
	Finish int `gorm:"column:finish"`
}
func (FileDO) TableName() string {
	return "file"
}


// table clients
type ClientDO struct {
	Id int `gorm:"column:uuid;primary_key"`
	LastRegTime *time.Time `gorm:"column:last_reg_time;type:DATETIME"`
}
func (ClientDO) TableName() string {
	return "client"
}

// table part
type PartDO struct {
	Id int `gorm:"column:id;auto_increment;primary_key"`
	Md5 string `gorm:"column:md5"`
	Size int64 `gorm:"column:size"`
}
func (PartDO) TableName() string {
	return "part"
}

// table relation_file_part
type FilePartRelationDO struct {
	Id int `gorm:"column:id;auto_increment;primary_key"`
	FileId int `gorm:"column:fid"`
	PartId int `gorm:"column:pid"`
}
func (FilePartRelationDO) TableName() string {
	return "relation_file_part"
}


// table sys
type SysDO struct {
	Key string `gorm:"column:key;primary_key"`
	Value string `gorm:"column:value"`
}
func (SysDO) TableName() string {
	return "sys"
}


// table tracker
type TrackerDO struct {
	Uuid string `gorm:"column:uuid;primary_key"`
	TrackerSyncId int `gorm:"column:tracker_sync_id"`
	LastRegTime *time.Time `gorm:"column:last_reg_time"`
	LocalPushId int `gorm:"column:local_push_id"`
}
func (TrackerDO) TableName() string {
	return "tracker"
}


// table web_storage_log
type WebStorageLogsDO struct {
	Id int `gorm:"column:id;auto_increment;primary_key"`
	StorageId int `gorm:"column:storage"`
	LogTime int64 `gorm:"column:log_time"`
	IOin int64 `gorm:"column:ioin"`
	IOout int64 `gorm:"column:ioout"`
	Disk int64 `gorm:"column:disk"`
	Memory int64 `gorm:"column:mem"`
	Download int `gorm:"column:download"`
	Upload int `gorm:"column:upload"`
}
func (WebStorageLogsDO) TableName() string {
	return "web_storage_log"
}


// table web_storage
type WebStorageDO struct {
	Id int `gorm:"column:id;auto_increment;primary_key"`
	Host string `gorm:"column:host"`
	Port int `gorm:"column:port"`
	Status int `gorm:"column:status"`
	TrackerId int `gorm:"column:tracker"`
	Uuid string `gorm:"column:uuid"`
	TotalFiles int `gorm:"column:total_files"`
	Group string `gorm:"column:grop"`
	InstanceId string `gorm:"column:instance_id"`
	HttpPort int `gorm:"column:http_port"`
	HttpEnable bool `gorm:"column:http_enable"`
	IOin int64 `gorm:"column:ioin"`
	IOout int64 `gorm:"column:ioout"`
	Disk int64 `gorm:"column:disk"`
	StartTime int64 `gorm:"column:start_time"`
	Download int `gorm:"column:downloads"`
	Upload int `gorm:"column:uploads"`
	ReadOnly int `gorm:"column:read_only"`
	Finish int `gorm:"column:finish"`
}
func (WebStorageDO) TableName() string {
	return "web_storage"
}


// table web_tracker
type WebTrackerDO struct {
	Id int `gorm:"column:id;auto_increment;primary_key"`
	Host string `gorm:"column:host"`
	Port int `gorm:"column:port"`
	Status int `gorm:"column:status"`
	Uuid string `gorm:"column:uuid"`
	TotalFiles int `gorm:"column:files"`
	Remark string `gorm:"column:remark"`
	Secret string `gorm:"column:secret"`
}
func (WebTrackerDO) TableName() string {
	return "web_tracker"
}

