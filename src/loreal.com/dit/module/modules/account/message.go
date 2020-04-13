package account

import (
	"errors"
	"log"

	"loreal.com/dit/module"
)

func (m *Module) registerHandlers() {

	m.AddMessageHandler("reload", func(msgPtr *module.Message) (handled bool) {
		m.disposeDB()
		m.initDB()
		log.Println("User account module reloaded!")
		return true
	})

	//Args[0] = token
	m.AddMessageHandler("getUserAccountByToken", func(msgPtr *module.Message) (handled bool) {
		token, ok := msgPtr.Args[0].(string)
		if !ok {
			msgPtr.Err = ErrInvalidParameter
			return true
		}
		userAccount := m.VerifyToken(token)
		if userAccount == nil {
			msgPtr.Err = errors.New("Invalid Token")
		}
		if msgPtr.ResultChan != nil {
			msgPtr.ResultChan <- userAccount
		}
		return true
	})

	//Args[0] = uid, Args[1] = password
	m.AddMessageHandler("getUserAccountByPassword", func(msgPtr *module.Message) (handled bool) {
		var uid, password string
		var ok bool
		uid, ok = msgPtr.Args[0].(string)
		if !ok {
			msgPtr.Err = ErrInvalidParameter
			return true
		}
		password, ok = msgPtr.Args[1].(string)
		if !ok {
			msgPtr.Err = ErrInvalidParameter
			return true
		}

		userAccount, loginErr := m.Authenticate(uid, password)
		msgPtr.Err = loginErr
		if msgPtr.ResultChan != nil {
			msgPtr.ResultChan <- userAccount
		}
		return true
	})

}
