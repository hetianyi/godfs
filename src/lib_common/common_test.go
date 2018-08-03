package lib_common

import (
    "testing"
    "fmt"
    "math"
)

func Test1(t *testing.T) {
    fmt.Println(FixLength(21211, 5, "0"))
    fmt.Println(int(math.Floor(12312300000*100*1.0/91231231234)))
}


func Test2(t *testing.T) {
    var a int64 = 9819391231
    println(fmt.Sprintf("%.2f", 2.345345))
    println(fmt.Sprintf("%.2f", float64(a)*1.0  * 1000))
}


