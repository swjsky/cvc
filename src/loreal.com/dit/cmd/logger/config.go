package main

import "strings"

// Configuration app configuration
type Configuration struct {
	Address        string `json:"address,omitempty"`
	RPCAddress     string `json:"rpc_address,omitempty"`
	MongoConnStr   string `json:"mongo_connection,omitempty"`
	MongoDBName    string `json:"mongo_dbname,omitempty"`
	SqliteDBName   string `json:"sqlite"`
	Prefix         string `json:"prefix,omitempty"`
	RedisServerStr string `json:"redis"`
}

func (c *Configuration) fixPrefix() {
	if !strings.HasPrefix(c.Prefix, "/") {
		c.Prefix = "/" + c.Prefix
	}
	if !strings.HasSuffix(c.Prefix, "/") {
		c.Prefix = c.Prefix + "/"
	}
}
