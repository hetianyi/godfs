package http

import (
    "net/http"
    "bytes"
)

func GetResponseBodyContent(resp *http.Response) (c string, e error) {
    body := resp.Body
    defer body.Close()
    bs := make([]byte, 10240)
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
