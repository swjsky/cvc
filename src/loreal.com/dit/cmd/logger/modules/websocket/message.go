package websocket

import (
	"loreal.com/dit/module"
	"log"
	"time"
)

func (m *Module) registerMessageHandlers() {
	m.AddMessageHandler("websocket-send", func(msgPtr *module.Message) bool {
		defer func() {
			if msgPtr.ResultChan != nil && msgPtr.Err != nil {
				log.Println(msgPtr.Err.Error())
				msgPtr.ResultChan <- false
			}
		}()

		notification, ok := msgPtr.Args[0].(Notification)
		if !ok {
			log.Println(ErrInvalidParameter.Error())
			msgPtr.Err = ErrInvalidParameter
			return true
		}

		m.WebsocketManager.ToSend <- notification

		timeout := time.NewTimer(3 * time.Second)
		sign := make(chan struct{})

		go func() {
			for {
				select {
				case <-m.WebsocketManager.ToSkip:
					msgPtr.ResultChan <- true
					sign <- struct{}{}
					return
				case <-timeout.C:
					msgPtr.ResultChan <- false
					msgPtr.Err = ErrReceiveTimeOut
					sign <- struct{}{}
					return
				}
			}
		}()

		<-sign

		log.Println("Websocket communicate finish")

		if msgPtr.ResultChan == nil {
			msgPtr.ResultChan <- false
		}

		return true
	})
}
