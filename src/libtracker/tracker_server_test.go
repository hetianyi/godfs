package libtracker

import (
	"libcommon"
	"testing"
	"util/db"
	"util/logger"
	"app"
	"libservicev2"
	"libcommon/bridgev2"
)

func init() {
	logger.SetLogLevel(1)
	app.SECRET = "123456"
	app.BASE_PATH = "E:\\godfs-storage\\storage1"
	app.PORT = 1022
	libservicev2.SetPool(db.NewPool(1))
}


func TestStartTrackerServer(t *testing.T) {
	app.UUID = "tracker01"
	server := bridgev2.NewServer("", app.PORT)
	server.Listen(libcommon.FutureExpireStorageServer)
}


