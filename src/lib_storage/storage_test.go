package lib_storage

import (
    "testing"
    "errors"
    "util/logger"
    "encoding/json"
    "fmt"
    "encoding/binary"
    "time"
    "crypto/md5"
    "encoding/hex"
    "util/common"
    "regexp"
    "container/list"
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


func Test6(t *testing.T) {
    pathRegexRestful := "^/download/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})(/([^/]*))?$"
    fmt.Println(regexp.Match(pathRegexRestful, []byte("/download/G01/01/432597de0e65eedbc867620e744a35ad")))
    fmt.Println(regexp.MustCompile(pathRegexRestful).ReplaceAllString("/download/G01/01/432597de0e65eedbc867620e744a35ad/12==-=-=*/", "${5}"))
    fmt.Println(time.Now().UTC().String())
}


func Test7(t *testing.T) {

    //Wed, 18 Jul 2018 04:49:08 GMT

    gmtLocation, _ := time.LoadLocation("GMT")
    fmt.Println(time.Now().In(gmtLocation).Format(time.RFC1123))

}


func Test8(t *testing.T) {
    var ls list.List
    ls.PushBack("xxx")
    fmt.Println(json.Marshal(ls))
}
func Test9(t *testing.T) {
    var m = make(map[string] int)
    fmt.Println(m["aaaa"])
}

func Test10(t *testing.T) {
    a := "^bytes=([0-9]+)-([0-9]+)?$"
    fmt.Println(regexp.Match(a, []byte("bytes=0-801")))
    fmt.Println(regexp.Match(a, []byte("bytes=0-45454545")))

    fmt.Println(parseHeaderRange("bytes=0-801"))
    fmt.Println(parseHeaderRange("bytes=0-45454545"))
    fmt.Println(parseHeaderRange("bytes=120-"))
    fmt.Println(parseHeaderRange("bytes=-"))
}
func Test11(t *testing.T) {
    a := "^multipart/form-data; boundary=.*$"
    fmt.Println(regexp.Match(a, []byte("multipart/form-data; boundary=------------------------fe43cbff9d519997")))
}
