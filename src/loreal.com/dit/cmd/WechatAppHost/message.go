package main

import (
	"log"

	"loreal.com/dit/module"
	"loreal.com/dit/utils"
)

func (a *App) initMessageHandlers() {
	a.MessageHandlers = map[string]func(*module.Message) bool{
		"reload": a.reloadMessageHandler,
	}

}

//reloadMessageHandler - handle reload message
func (a *App) reloadMessageHandler(msgPtr *module.Message) (handled bool) {
	//reload configuration
	utils.LoadOrCreateJSON("./config/config.json", &a.Config)
	a.Config.fixPrefix()
	a.StartScheduler()
	log.Println("[INFO] - Configuration reloaded!")
	return true
}
