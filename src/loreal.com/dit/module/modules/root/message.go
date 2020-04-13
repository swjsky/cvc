package root

import (
	"log"

	"loreal.com/dit/module"
)

func (m *Module) registerHandlers() {

	m.AddMessageHandler("reload", func(msgPtr *module.Message) (handled bool) {
		reloadAccounts()
		log.Println("Root module reloaded!")
		return true
	})

	// //Args[0] = mobile, Args[1] = token
	// m.AddMessageHandler("verify-token", func(msgPtr *module.Message) (handled bool) {
	// 	mobile := msgPtr.Args[0].(string)
	// 	token := msgPtr.Args[1].(string)
	// 	if m.verifyToken(mobile, token) {
	// 		msgPtr.Err = nil
	// 		msgPtr.Results = make([]interface{}, 1)
	// 		msgPtr.Results[0] = true
	// 		msgPtr.Callback(msgPtr)
	// 		return true
	// 	}
	// 	msgPtr.Err = nil
	// 	msgPtr.Results = make([]interface{}, 1)
	// 	msgPtr.Results[0] = false
	// 	msgPtr.Callback(msgPtr)
	// 	return true
	// })

}
