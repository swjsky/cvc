package main

import (
	"log"

	"loreal.com/dit/module"
	"loreal.com/dit/module/modules/root"
	"loreal.com/dit/utils"
)

func registerHandlers(m *root.Module) {
	m.AddMessageHandler("reload", func(msgPtr *module.Message) (handled bool) {
		//reload configuration
		utils.LoadOrCreateJSON("./config/config.json", &Cfg)
		log.Println("Configuration reloaded!")
		log.Println(*Cfg)
		return true
	})
}
