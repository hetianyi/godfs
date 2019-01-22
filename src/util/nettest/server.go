package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("listening...")
	listener, err := net.Listen("tcp", ":8799")
	if err != nil {
		fmt.Println("error receive conn")
		return
	}
	// keep accept connections.
	conn, _ := listener.Accept()
	//-----------------------------------

	fmt.Println(conn.RemoteAddr().String())



}
