package grom

import (
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "github.com/jinzhu/gorm"
)

type PartDO struct {
    Id int `gorm:"column:id;auto_increment;primary_key"`
    Md5 string `gorm:"column:md5"`
    Size uint `gorm:"column:size"`
}
// Set User's table name to be `profiles`
func (PartDO) TableName() string {
    return "parts"
}


func InsertFileDO(partDO *PartDO) {
    db, err := gorm.Open("sqlite3", "E:/godfs-storage/storage1/data/storage.db")
    if err != nil {
        panic("failed to connect database")
    }
    defer db.Close()
    db.LogMode(true)

    // Create
    db.Create(&PartDO{Md5: "981723324abcdff", Size: 1000})

    // Read
    var product PartDO
    db.First(&product, 62) // find product with id 1
    db.First(&product, "md5 = ?", "981723324abcdff") // find product with code l1212

    // Update - update product's price to 2000
    db.Model(&product).Update("Size", 123456)

    // Delete - delete product
    db.Delete(&product)

}