package main

import (
	"strings"

	"loreal.com/dit/utils"
)

//Config - data struct for configuration file
type Config struct {
	Address  string `json:"address"`
	Prefix   string `json:"prefix"`
	BasePath string `json:"base_path"`
}

var userAccounts map[string]string

func (c *Config) fixPrefix() {
	if !strings.HasPrefix(c.Prefix, "/") {
		c.Prefix = "/" + c.Prefix
	}
	if !strings.HasSuffix(c.Prefix, "/") {
		c.Prefix = c.Prefix + "/"
	}
}

func loadDefaultUser() {
	userAccounts = map[string]string{
		"loreal": "Aa123456",
	}
	utils.LoadOrCreateJSON("./config/account.json", userAccounts)
}

func init() {
	loadDefaultUser()
}
