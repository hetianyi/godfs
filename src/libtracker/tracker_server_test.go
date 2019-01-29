package libtracker

import (
	"app"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
	"testing"
	"util/db"
	"util/logger"
)

func init() {
	logger.SetLogLevel(1)
	app.Secret = "123456"
	app.BasePath = "E:\\godfs-storage\\storage1"
	app.Port = 1022
	libservicev2.SetPool(db.NewPool(1))
}

func TestStartTrackerServer(t *testing.T) {
	app.UUID = "tracker01"
	server := bridgev2.NewServer("", app.Port)
	server.Listen(libcommon.FutureExpireStorageServer)
}
