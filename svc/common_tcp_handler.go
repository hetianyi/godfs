package svc

import "github.com/hetianyi/godfs/api"

const MaxConnPerServer uint = 100

var clientAPI api.ClientAPI

// InitializeClientAPI initializes client API.
func InitializeClientAPI(config *api.Config) {
	clientAPI = api.NewClient()
	clientAPI.SetConfig(config)
}
