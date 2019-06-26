package api

import (
	"container/list"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/conn"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"io"
	"net"
	"sync"
	"time"
)

const (
	DefaultMaxConnectionsPerServer = 100
)

var (
	NoStorageServerErr = errors.New("no storage available")
)

type Config struct {
	MaxConnectionsPerServer  uint                   // limit max connection for each server
	TrackerServers           []common.Server        // tracker servers
	StaticStorageServers     []common.StorageServer // storage servers
	RegisteredStorageServers []common.StorageServer // storage servers
	// Trackers or Storages
}

type ClientAPI interface {
	// Init initializes the ClientAPI.
	Init(config *Config) error
	// RefreshConfig(config *Config) // TODO
	Upload(src io.Reader, group string) (bool, error)
	Download(input io.Reader) (bool, error)
}

type clientAPIImpl struct {
	config  *Config
	lock    *sync.Mutex
	weights map[string]int64 // server use weights
}

func (c *clientAPIImpl) Init(config *Config) {
	if config != nil {
		c.config = config
	} else {
		c.config = &Config{
			MaxConnectionsPerServer: DefaultMaxConnectionsPerServer,
		}
	}

	if c.config.MaxConnectionsPerServer <= 0 {
		c.config.MaxConnectionsPerServer = DefaultMaxConnectionsPerServer
	}
	if (c.config.TrackerServers == nil || len(c.config.TrackerServers) == 0) &&
		(c.config.StaticStorageServers == nil || len(c.config.StaticStorageServers) == 0) {
		logger.Warn("client initialized but no server provided")
	}
	if c.config.TrackerServers != nil {
		for _, s := range c.config.TrackerServers {
			conn.InitServerSettings(&conn.Server{
				Host: s.Host,
				Port: s.Port,
			}, c.config.MaxConnectionsPerServer, time.Minute*5)
		}
	}
	if c.config.StaticStorageServers != nil {
		for _, s := range c.config.StaticStorageServers {
			conn.InitServerSettings(&conn.Server{
				Host: s.Host,
				Port: s.Port,
			}, c.config.MaxConnectionsPerServer, time.Minute*5)
		}
	}
}

func (c *clientAPIImpl) Upload(src io.Reader, group string) (bool, error) {
	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error

	for true {
		selectedStorage = c.selectStorageServer(group, exclude)
		if selectedStorage == nil {
			return false, NoStorageServerErr
		}
		connection, err := conn.GetConnection(selectedStorage.ToServer())
		if err != nil {
			lastErr = err
			exclude.PushBack(selectedStorage)
			continue
		}
		pip := &gpip.Pip{
			Conn: conn,
		}

	}

	return false, nil
}

// selectStorageServer selects proper storage server.
func (c *clientAPIImpl) selectStorageServer(group string, exclude *list.List) *common.StorageServer {
	c.lock.Lock()
	defer c.lock.Unlock()
	var candidates = list.New()
	// if registered storage server is not empty, use it first.
	if c.config.RegisteredStorageServers != nil {
		for _, s := range c.config.RegisteredStorageServers {
			if isExcluded(&s, exclude) {
				continue
			}
			candidates.PushBack(s)
		}
	}
	if candidates.Len() == 0 {
		for _, s := range c.config.StaticStorageServers {
			if isExcluded(&s, exclude) {
				continue
			}
			candidates.PushBack(s)
		}
	}
	// select smallest weights of storage server.
	var selectedStorage *common.StorageServer
	gox.WalkList(candidates, func(item interface{}) bool {
		if selectedStorage == nil {
			selectedStorage = item.(*common.StorageServer)
			return false
		}
		if c.weights[item.(*common.StorageServer).InstanceId] < c.weights[selectedStorage.InstanceId] {
			selectedStorage = item.(*common.StorageServer)
			return false
		}
		return false
	})
	return selectedStorage
}

func isExcluded(s *common.StorageServer, exclude *list.List) bool {
	con := false
	if exclude != nil {
		gox.WalkList(exclude, func(item interface{}) bool {
			if item.(*common.StorageServer).InstanceId == s.InstanceId {
				con = true
				return true
			}
			return false
		})
	}
	return con
}
