package bridgev2

import (
	"testing"
	"app"
	"fmt"
	"util/json"
	"util/logger"
	"libservicev2"
)

func init() {
	logger.SetLogLevel(1)
	app.SECRET = "123456"
	app.BASE_PATH = "E:\\godfs-storage\\storage1"
	libservicev2.SetPool(libservicev2.NewPool(1))
}

func PrintResult(result... interface{}) {
	fmt.Println("\n\n+++~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~+++")
	if result != nil {
		for i := range result {
			obj := result[i]
			bs, _ := json.Marshal(obj)
			fmt.Println(string(bs))
		}
	}
	fmt.Println("+++~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~+++")
}

func TestServer(t *testing.T) {
	app.UUID = "tracker01"
	server := NewServer("", 1022)
	server.Listen()
}



func TestClient(t *testing.T) {
	app.UUID = "storage01"
	member := &app.Member{
		LookBackAddress: "localhost",
		Port: 1022,
		AdvertiseAddr: "192.168.1.142",
		AdvertisePort: 1022,
		AccessFlag: app.ACCESS_FLAG_NONE,
	}
	server := &app.ServerInfo{}
	server.FromMember(member)
	index := 0
	for {
		client := NewClient(server)
		if err := client.Connect(); err != nil {
			panic(err)
		}
		index++
		PrintResult(client.Validate())
		client.Destroy()
		fmt.Println(index)
	}

}