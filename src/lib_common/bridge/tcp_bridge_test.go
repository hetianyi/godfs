package bridge

import (
    "testing"
    "fmt"
    "encoding/json"
)


func add(closer SendReceiveCloser) {
    fmt.Println(json.Marshal(closer))
}

func Test1(t *testing.T) {
    req := &Request{}
    add(req)
}
