package badger_test

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("D:/my", 0600, nil)
	if err != nil {
		logger.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	fmt.Println("开始插入：" + gox.GetLongLongDateString(time.Now()))
	for i := 0; i < 100000; i++ {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("MyBucket"))
			b.Put([]byte("name"+convert.IntToStr(i)), []byte("A3AE1i_kNI5gneeop4tWUocv9bYLwyiXDuJSker1VmeWWJ0ioeLA6jIWyPrtRmsZo3-fQl32vL7DZF9Gtc1wcw"))
			logger.Info("插入：", "name"+convert.IntToStr(i))
			return nil
		})
	}
	fmt.Println("结束插入：" + gox.GetLongLongDateString(time.Now()))
	db.Close()
}

func Test2(t *testing.T) {

	db, err := bolt.Open("D:/my", 0600, nil)
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println("开始查询：" + gox.GetLongLongDateString(time.Now()))
	for i := 0; i < 1000; i++ {
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("MyBucket"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			ret := b.Get([]byte("name" + convert.IntToStr(i)))
			logger.Info("查询到："+"name"+convert.IntToStr(i), " --> ", string(ret))
			return nil
		})

	}
	fmt.Println("结束查询：" + gox.GetLongLongDateString(time.Now()))

	fmt.Println("结束")
}
