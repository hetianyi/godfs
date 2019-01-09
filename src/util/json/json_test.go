package json

import (
    "testing"
    "fmt"
)

type User struct {
    Name string `json:"name"`
}

func TestMarshal(t *testing.T) {
    m := make(map[string]string)
    m["name"] = "xxx"
    bs, _ := Marshal(&m)
    fmt.Println(string(bs))
}

func TestUnmarshal(t *testing.T) {
    a := "{\"name\":\"xxx\"}"
    m :=&User{}
    UnmarshalFromString(a, &m)
    fmt.Println(m.Name)
}