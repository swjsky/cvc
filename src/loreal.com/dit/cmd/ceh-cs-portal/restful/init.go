package restful

import "os"

//DEBUG - whether in debug mode
var DEBUG bool

//INFOLEVEL - info level for debug mode
var INFOLEVEL int

//LOGLEVEL - info level for logs
var LOGLEVEL int

func init() {
	if os.Getenv("EV_DEBUG") != "" {
		DEBUG = true
	}
	INFOLEVEL = 1
	LOGLEVEL = 1
}
