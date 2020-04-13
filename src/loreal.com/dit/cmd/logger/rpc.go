package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"loreal.com/dit/cmd/logger/modle"
)

// Send rpc send
func (app *App) Send(req *modle.Logger, reply *string) error {
	// Create log
	if err := app.Create(req); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = "ok"
	// Send to ws
	resultChan := make(chan interface{}, 1)
	app.Root.SendWithChan("websocket-send", resultChan, req)
	if result, ok := <-resultChan; ok {
		if sendResult := result.(bool); !sendResult {
			log.Println("Send to ws error")
		}
	}
	return nil
}

// Ping rpc ping
func (app *App) Ping(req interface{}, reply *string) error {
	*reply = "ok"
	return nil
}

// Start start rpc
func (app *App) Start() {
	addr := app.Config.RPCAddress
	server := rpc.NewServer()
	server.Register(app)

	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatalln("listen occur error", e)
	} else {
		log.Println("listen on..", addr)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("listener accept error:", err)
			continue
		}
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
