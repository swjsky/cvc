//Package root - The root module for all CEH application
//It implements the basic infrastructure for application lifecycle management : start/stop/restart
//As well as instrumentation infrastructure through using prometheus
package root

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"log"

	"loreal.com/dit/module"
	"loreal.com/dit/registry/consul"
	"loreal.com/dit/utils"
)

//Module - Micro service module for root path
type Module struct {
	*module.Module
	Context    map[string]interface{}
	HTTPServer *http.Server
	Addr       string
	CertFile   string
	KeyFile    string
	Peers      []*Peer
	Mutex      *sync.RWMutex
}

//GetBase - implements ISubModule interface
func (m *Module) GetBase() *module.Module {
	return m.Module
}

func fixPrefix(prefix string) string {
	if prefix != "/" {
		return strings.TrimRight("/"+strings.Trim(prefix, "/"), "/")
	}
	return ""
}

//NewModule create root module
func NewModule(name, desc, prefix string) *Module {
	m := &Module{
		Module:  module.NewModule(name, desc, ""),
		Context: make(map[string]interface{}),
		Mutex:   &sync.RWMutex{},
	}
	m.Prefix = fixPrefix(prefix)
	m.Path = ""
	m.NameSpace = ""
	m.Peers = make([]*Peer, 0)

	m.OnShutdown = m.onShutdown

	m.registerHandlers()
	m.registerEndpoints()
	return m
}

//NewConsulRegisterModule create root module
func NewConsulRegisterModule(name, desc, prefix, addr string) *Module {
	registryBackend := &consul.Consul{
		Addr:          "localhost:8500",
		Scheme:        "http",
		KVPath:        "/ceh/config",
		TagPrefix:     "urlprefix-",
		Register:      true,
		ServiceAddr:   ":9998",
		ServiceName:   "ceh",
		ServiceStatus: []string{"passing"},
		CheckInterval: time.Second,
		CheckTimeout:  3 * time.Second,
	}

	utils.LoadOrCreateJSON("./config/consul.json", &registryBackend)

	m := NewModule(name, desc, prefix)
	m.Addr = addr
	regrstry, err := consul.NewRegistry(registryBackend)
	if err != nil {
		//skip consul registration
		log.Println("[ERR] - [Consul]: skipping service registration")
		return m
	}
	unRegSignal := regrstry.Register(name, desc, addr, prefix)
	//Unregister on dispose
	m.OnDispose = func(p *module.Module) {
		regrstry.Unregister(unRegSignal)
	}
	return m
}

//GetContext - get data from context by key
func (m *Module) GetContext(key string) (value interface{}) {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	value = m.Context[key]
	return
}

//SetContext - set data into context
func (m *Module) SetContext(key string, value interface{}) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Context[key] = value
}

//defaultHTTPServer - create HTTP server with default parameters
func (m *Module) defaultHTTPServer(addr string) *http.Server {
	m.Addr = addr
	m.HTTPServer = &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 1<<63 - 1, //TODO: 另外单独配合各服务器做长轮询，否则所有的http请求都是这样。
	}
	m.HTTPServer.SetKeepAlivesEnabled(true)
	m.HTTPServer.IdleTimeout = time.Second * 3
	return m.HTTPServer
}

func (m *Module) onShutdown(p *module.Module) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	defer cancel()
	if m.HTTPServer != nil {
		log.Println("[INFO] - Graceful shutdown...")
		// if err := m.HTTPServer.Close(); err != nil {
		// 	log.Println("[ERR] - Graceful shutdown:", err)
		// }
		if err := m.HTTPServer.Shutdown(ctx); err != nil {
			log.Println("[WARNING] - Graceful shutdown:", err)
		}
	}
	m.HTTPServer = nil
}

//ListenAndServe - Helper func to start HTTP service for root module
func (m *Module) ListenAndServe(addr string) {
	m.defaultHTTPServer(addr)
	m.Start()
	go signalListen(m)

	log.Printf("[INFO] - CEH [%s]: Listening on %s", m.Description, m.Addr)

	if m.CertFile != "" && m.KeyFile != "" {
		if err := m.HTTPServer.ListenAndServeTLS(m.CertFile, m.KeyFile); err != nil {
			log.Fatal("[ERR] - ListenAndServeTLS:", err)
		}
		return
	}
	if err := m.HTTPServer.ListenAndServe(); err != nil {
		log.Fatal("[ERR] - ListenAndServe:", err)
	}
}

func signalListen(root *Module) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	for {
		<-c
		root.Shutdown()
	}
}

// disposeDB module resources
// func (m *Module) disposeDB() {
// 	if m.db != nil {
// 		m.db.Close()
// 	}
// }

// func (m *Module) initDB() {
// 	var err error
// 	m.db, err = sql.Open("sqlite3", "./sms.db?cache=shared&mode=rwc")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	//init database tables
// 	sqlStmts := []string{
// 		"create table if not exists smscode (mobile int not null primary key,sentCount int default 1,refCode text, smsCode text,sentAt datetime);",
// 		"create index if not exists smscodeidx on smscode(mobile);",
// 	}
// 	for _, sqlStmt := range sqlStmts {
// 		_, err := m.db.Exec(sqlStmt)
// 		if err != nil {
// 			log.Printf("%q: %s\n", err, sqlStmt)
// 			return
// 		}
// 	}
// }
