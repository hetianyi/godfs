package storage_test

import (
    "testing"
    "common"
    "errors"
    "util/logger"
    "encoding/json"
    "fmt"
)


func dev(a int, b int) (int, error) {
    return 0, errors.New("xxxxx")
}

func Test1(t *testing.T) {
    /*a := "aaa,"
    println(len(strings.Split(a, ",")))*/
    /*bool, _ := regexp.Match("^[1-9][0-9]{1,6}$", []byte("190000"))
    println(bool)*/

    common.Try(func() {
        r, e := dev(1, 0)
        if e != nil {
            logger.Debug("panic...")
            panic(e)
        } else {
            logger.Info("the result is :" , r)
        }
    }, func(i interface{}) {
        logger.Fatal("get error : ", i)
    })
}


type Man struct {
    Age int `json:"age1"`
}

func Test2(t *testing.T) {
    a := "{\"age\": 23}"
    var man Man
    bs1, _ := json.Marshal(man)
    println(string(bs1))

    json.Unmarshal([]byte(a), &man)
    bs, e := json.Marshal(man)
    println(string(bs))
    fmt.Println(e)
}





