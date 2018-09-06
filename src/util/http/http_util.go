package http

import (
    "net/http"
    "bytes"
    "lib_common/bridge"
)

func GetResponseBodyContent(resp *http.Response) (c string, e error) {
    body := resp.Body
    defer body.Close()
    bs, _ := bridge.MakeBytes(10240, false, 0)
    var buffer bytes.Buffer
    for {
        len, err := body.Read(bs)
        if err == nil {
            buffer.Write(bs[0:len])
        } else {
            defer func() {
                e = err
            }()
            break
        }
    }
    return buffer.String(), e
}
