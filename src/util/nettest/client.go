package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	for {
		conn, e := net.Dial("tcp", "localhost:8799")
		if e != nil {
			fmt.Println("error connect to server")
			continue
		}
		conn.Close()
		time.Sleep(1000)
	}
}
