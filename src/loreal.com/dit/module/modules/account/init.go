package account

import (
	"log"
	"os"
)

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

func debugInfo(source, msg string, level int) {
	if DEBUG && INFOLEVEL >= level {
		log.Printf("[DEBUG] - [%s]%s\n", source, msg)
	}
}
