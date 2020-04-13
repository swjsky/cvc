package websocket

import (
	"loreal.com/dit/module"
)

const (
	moduleName        = "websocket"
	moduleDescription = "revo websocket"
)

// Module websocket module
type Module struct {
	*module.Module
	WebsocketManager *Manager
}

// GetBase return module moudle
func (m *Module) GetBase() *module.Module {
	return m.Module
}

// NewModule return websocket module
func NewModule(path string) *Module {
	m := &Module{
		Module: module.NewModule(moduleName, moduleDescription, path),
	}

	m.registerMessageHandlers()
	m.registerEndpoints()

	wsm := NewWebsocketManager()
	m.WebsocketManager = wsm
	go wsm.Listen()

	m.OnDispose = func(p *module.Module) {
		wsm.Close()
	}
	return m
}
