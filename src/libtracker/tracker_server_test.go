package libtracker

import (
	"testing"
	"util/logger"
	"app"
	"libservicev2"
	"libcommon/bridgev2"
)

func init() {
	logger.SetLogLevel(1)
	app.SECRET = "123456"
	app.BASE_PATH = "E:\\godfs-storage\\storage1"
	libservicev2.SetPool(libservicev2.NewPool(1))
}


func TestStartTrackerServer(t *testing.T) {
	app.UUID = "tracker01"
	server := bridgev2.NewServer("", 1022)
	server.Listen()
}


