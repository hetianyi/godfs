package libcommon

import "time"

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
// Set User's table name to be `profiles`
func (FileDO) TableName() string {
    return "files"
}


// table clients
type ClientDO struct {
    Id int `gorm:"column:uuid;primary_key"`
    LastRegTime *time.Time `gorm:"column:last_reg_time;type:DATETIME"`
}
// Set User's table name to be `profiles`
func (ClientDO) TableName() string {
    return "clients"
}

// table clients
type PartDO struct {
    Id int `gorm:"column:id;auto_increment;primary_key"`
    Md5 string `gorm:"column:md5"`
    Size int64 `gorm:"column:size"`
}
// Set User's table name to be `profiles`
func (PartDO) TableName() string {
    return "parts"
}

// table clients
type PartRelationDO struct {
    Id int `gorm:"column:id;auto_increment;primary_key"`
    FileId int `gorm:"column:fid"`
    PartId int `gorm:"column:pid"`
}
// Set User's table name to be `profiles`
func (PartRelationDO) TableName() string {
    return "parts_relation"
}


// table clients
type SysDO struct {
    Key string `gorm:"column:key;primary_key"`
    Value string `gorm:"column:value"`
}
// Set User's table name to be `profiles`
func (SysDO) TableName() string {
    return "sys"
}


// table clients
type TrackerDO struct {
    Uuid string `gorm:"column:uuid;primary_key"`
    TrackerSyncId int `gorm:"column:tracker_sync_id"`
    LastRegTime *time.Time `gorm:"column:last_reg_time"`
    LocalPushId int `gorm:"column:local_push_id"`
}
// Set User's table name to be `profiles`
func (TrackerDO) TableName() string {
    return "trackers"
}


