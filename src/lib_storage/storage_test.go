package lib_storage

import (
    "testing"
    "common"
    "errors"
    "util/logger"
    "encoding/json"
    "fmt"
    "encoding/binary"
    "time"
    "crypto/md5"
    "encoding/hex"
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


func Test3(t *testing.T) {
    //a := []byte{1,1,0,1}
    i := uint64(1024*1024*1024*1024*1024*1024*15)
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, i)
    fmt.Println(b[:])

    c := binary.BigEndian.Uint64(b)
    fmt.Println(c)

    a1 := "12345a"
    b1 := "12345a"

    fmt.Println(a1 == b1)


    a2 := make(map[string] string)

    a2["name"] = "lisi"
    a2["name1"] = ""

    fmt.Println(a2["name"])
    fmt.Println(a2["name1"]=="")

}



func Test4(t *testing.T) {
    timer := time.NewTicker(time.Second * 1)
    for {
        <-timer.C
        fmt.Println("x")
    }

}


func Test5(t *testing.T) {
    a := []byte{1,2,3,4,5,6,7}
    md := md5.New()
    md.Write(a)
    cipherStr := md.Sum(nil)
    fmt.Println(hex.EncodeToString(cipherStr))
}
