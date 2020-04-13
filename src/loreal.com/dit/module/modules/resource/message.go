package resource

import (
	"log"

	"loreal.com/dit/module"
)

func (m *Module) registerHandlers() {

	m.AddMessageHandler("reload", func(msgPtr *module.Message) (handled bool) {
		log.Println("reloaded!")
		return true
	})

}
