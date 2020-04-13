//Package module define the base struct for modules
package module

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"

	"loreal.com/dit/endpoint"
)

const (
	//StatusNew - Module is new, hasn't initialized yet
	StatusNew = iota
	//StatusInitializing Module is initializing
	StatusInitializing
	//StatusInitialized Module is initialized
	StatusInitialized
	//StatusStarting - running
	StatusStarting
	//StatusRunning - Module is running
	StatusRunning
	//StatusStopping - Module is stopping
	StatusStopping
	//StatusStopped - Module is stopped
	StatusStopped
	//StatusDisposed - Module is dZisposed
	StatusDisposed
)

//Module - Micro service for CEH
type Module struct {
	ID              uuid.UUID
	Root            *Module
	Parent          *Module
	NameSpace       string
	Name            string
	Description     string
	Prefix          string
	Path            string
	Children        []ISubModule
	MessageBus      chan *Message
	MountingPoints  map[string]endpoint.Endpoint
	MessageHandlers map[string][]MessageHandler
	status          int32
	WatchDogTick    time.Duration
	OnTick          func(p *Module)
	OnInit          func(p *Module)
	OnDispose       func(p *Module)
	OnStart         func(p *Module)
	OnStop          func(p *Module)
	OnShutdown      func(p *Module)
	wg              *sync.WaitGroup
}

//NewModule - Create new micro service for CEH
func NewModule(name, description, path string) *Module {
	path = strings.Trim(path, "/")
	m := &Module{
		ID:              uuid.NewV4(),
		Name:            name,
		Description:     description,
		Path:            path,
		Children:        make([]ISubModule, 0),
		MountingPoints:  make(map[string]endpoint.Endpoint),
		MessageHandlers: make(map[string][]MessageHandler),
		MessageBus:      make(chan *Message, 0),
		status:          StatusNew,
		WatchDogTick:    15 * time.Minute,
		wg:              new(sync.WaitGroup),
	}
	m.SetRoot(m)
	return m
}

//SetRoot set root module
func (p *Module) SetRoot(root *Module) {
	p.Root = root
}

//SetParent set parent module
func (p *Module) SetParent(parent *Module) {
	p.Parent = parent
	p.SetPrefix(parent.Fullpath())
	p.SetNameSpace(parent.String())
}

//Install sub modules
func (p *Module) Install(subModules ...ISubModule) {
	for _, m := range subModules {
		p.Children = append(p.Children, m)
		if hasRoot, ok := m.(HasRoot); ok {
			if p.Root != nil {
				hasRoot.SetRoot(p.Root)
			} else {
				hasRoot.SetRoot(p)
			}
		}
		if hasParent, ok := m.(HasParent); ok {
			hasParent.SetParent(p)
		}
	}
}

//SetPrefix of current module
func (p *Module) SetPrefix(prefix string) {
	p.Prefix = "/" + strings.Trim(prefix, "/")
	p.Prefix = strings.TrimRight(prefix, "/")
}

//SetNameSpace of current module
func (p *Module) SetNameSpace(nameSpace string) {
	p.NameSpace = strings.Trim(nameSpace, ".")
}

func (p *Module) String() string {
	if p.NameSpace == "" {
		return p.Name
	}
	return fmt.Sprintf("%s.%s", p.NameSpace, p.Name)
}

//Fullpath return full path to module
func (p *Module) Fullpath() string {
	switch {
	case p.Prefix == "" && p.Path == "":
		return "/"
	case p.Prefix != "" && p.Path == "":
		return p.Prefix + "/"
	default:
		return fmt.Sprintf("%s/%s", p.Prefix, p.Path)
	}
}

//GetStatus read status of current module
func (p *Module) GetStatus() int32 {
	return atomic.LoadInt32(&p.status)
}

//SetStatus set status of current module
func (p *Module) SetStatus(status int32) {
	atomic.StoreInt32(&p.status, status)
}

//Init micro service module, inject the initFunc
func (p *Module) Init() {
	if p.GetStatus() == StatusNew {
		p.SetStatus(StatusInitializing)
		if p.OnInit != nil {
			p.OnInit(p)
		}
		p.Mount()
		p.SetStatus(StatusInitialized)
	}
}

//Mount endpoints for micro service module, implements mountable interface
func (p *Module) Mount() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[ERR] - Module:[%s] -> Mounting error: %s\r\n", p, err)
		}
	}()
	for ep, epHandler := range p.MountingPoints {
		switch {
		case p.Prefix == "" && p.Path == "" && ep == "":
			log.Printf("[INFO] - Module:[%s] -> Mounting: %s\r\n", p, "/")
			http.Handle("/", epHandler)
		case p.Path == "" && ep != "":
			log.Printf("[INFO] - Module:[%s] -> Mounting: %s\r\n", p, p.Fullpath()+ep)
			http.Handle(p.Fullpath()+ep, epHandler)
		case ep == "":
			log.Printf("[INFO] - Module:[%s] -> Mounting: %s\r\n", p, p.Fullpath())
			http.Handle(p.Fullpath(), epHandler)
		default:
			log.Printf("[INFO] - Module:[%s] -> Mounting: %s\r\n", p, p.Fullpath()+"/"+ep)
			http.Handle(p.Fullpath()+"/"+ep, epHandler)
		}
	}
}

//Start micro service module
func (p *Module) Start() {
	status := p.GetStatus()
	switch status {
	case StatusNew:
		p.Init()
	case StatusInitialized, StatusStopped:
	default:
		log.Printf("[INFO] - Module:[%s] -> Can not start!\r\n", p)
		return
	}
	p.SetStatus(StatusStarting)
	go p.startMessageLoop()
	for _, c := range p.Children {
		if startable, ok := c.(Startable); ok {
			startable.Start()
		}
	}
	if p.OnStart != nil {
		p.OnStart(p)
	}
	log.Printf("[INFO] - Module:[%s] -> Started\r\n", p)
	p.SetStatus(StatusRunning)
	return
}

//Stop micro service module
func (p *Module) Stop() {
	if p.GetStatus() != StatusRunning {
		return
	}
	p.SetStatus(StatusStopping)
	if p.OnStop != nil {
		p.OnStop(p)
	}
	p.OnTick = nil
	for _, c := range p.Children {
		if stoppable, ok := c.(Stoppable); ok {
			stoppable.Stop()
		}
	}
	p.wg.Wait()
	close(p.MessageBus)
	p.SetStatus(StatusStopped)
	log.Printf("[INFO] - Module:[%s] -> Stopped\r\n", p)
}

//Shutdown current module
func (p *Module) Shutdown() {
	log.Printf("[INFO] - Module:[%s] -> Shutting down\r\n", p)
	p.Stop()
	p.Dispose()
	if p.OnShutdown != nil {
		p.OnShutdown(p)
	}
	log.Printf("[INFO] - Exit\r\n")
	os.Exit(0)
}

//Restart current module
func (p *Module) Restart() {
	log.Printf("[INFO] - Module:[%s] -> Restarting\r\n", p)
	p.Stop()
	p.Start()
}

//Reload micro service module
func (p *Module) Reload() {
	if p.GetStatus() != StatusRunning {
		return
	}
	log.Printf("[INFO] - Module:[%s] -> Reload message sent to sub-modules\r\n", p)
	p.Broadcast("reload", nil)
}

//Dispose resources for micro service module
func (p *Module) Dispose() {
	for _, c := range p.Children {
		if disposable, ok := c.(Disposable); ok {
			disposable.Dispose()
		}
	}
	if p.OnDispose != nil {
		p.OnDispose(p)
	}
	p.SetStatus(StatusDisposed)
	log.Printf("[INFO] - Module:[%s] -> Disposed\r\n", p)
}
