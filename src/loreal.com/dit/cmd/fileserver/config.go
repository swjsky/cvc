package main

import "strings"

//Config - data struct for configuration file
type Config struct {
	Auth           bool   `json:"auth"`
	Address        string `json:"address"`
	Prefix         string `json:"prefix"`
	ShareFolder    string `json:"share-folder"`
	MongoConnStr   string `json:"mongo-connection"`
	RedisServerStr string `json:"redis-server"`
}

func (c *Config) fixPrefix() {
	if !strings.HasPrefix(c.Prefix, "/") {
		c.Prefix = "/" + c.Prefix
	}
	if !strings.HasSuffix(c.Prefix, "/") {
		c.Prefix = c.Prefix + "/"
	}
}
