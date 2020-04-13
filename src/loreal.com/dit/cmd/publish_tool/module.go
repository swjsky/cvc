package main

import (
	"loreal.com/dit/module"
	"sync"
)

//Module - Micro service module to send sms verification code
type Module struct {
	*module.Module
	mutex *sync.Mutex
}

//GetBase - Get base module
func (m *Module) GetBase() *module.Module {
	return m.Module
}

//NewModule create new sms module
func NewModule(path string) *Module {
	m := &Module{
		Module: module.NewModule("publish tool", "publish tool", path),
		mutex:  &sync.Mutex{},
	}
	m.registerEndpoints()

	return m
}
