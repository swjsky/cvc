package main

import (
	"loreal.com/dit/module"
	"loreal.com/dit/module/modules/root"
	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
	"loreal.com/dit/utils"
	"log"
	"net/http"
	"sync"

	"github.com/robfig/cron"
)

// App data struct for App & configuration file
type App struct {
	Name              string
	Description       string
	Config            *Configuration
	Root              *root.Module
	Endpoints         map[string]EndpointEntry
	MessageHandlers   map[string]func(*module.Message) bool
	AuthProvider      middlewares.RoleVerifier
	MgoSessionManager *utils.MongoSessionManager
	Scheduler         *cron.Cron
	mutex             *sync.Mutex
}

// EndpointEntry - endpoint registry entry
type EndpointEntry struct {
	Handler     func(http.ResponseWriter, *http.Request)
	Middlewares []endpoint.ServerMiddleware
}

// NewApp create new app
func NewApp(name, description string, config *Configuration) *App {
	if config == nil {
		log.Println("Missing configuration data")
		return nil
	}
	app := &App{
		Root:              root.NewModule(name, description, config.Prefix),
		Name:              name,
		Description:       description,
		Config:            config,
		MgoSessionManager: utils.NewMongoSessionManager(name, config.MongoConnStr, 200),
		Scheduler:         cron.New(),
		mutex:             &sync.Mutex{},
	}
	app.Init()
	return app
}

// Init app initialization
func (a *App) Init() {
	if a.Config != nil {
		a.Config.fixPrefix()
	}
	a.registerEndpoints()
	a.registerMessageHandlers()
}

// registerEndpoints - Register Endpoints
func (a *App) registerEndpoints() {
	a.initEndpoints()
	for path, entry := range a.Endpoints {
		if entry.Middlewares == nil {
			entry.Middlewares = a.getDefaultMiddlewares(path)
		}
		a.Root.MountingPoints[path] = endpoint.DecorateServer(
			endpoint.Impl(entry.Handler),
			entry.Middlewares...,
		)
	}
}

// registerMessageHandlers - Register Message Handlers
func (a *App) registerMessageHandlers() {
	a.initMessageHandlers()
	for path, handler := range a.MessageHandlers {
		a.Root.AddMessageHandler(path, handler)
	}
}

// ListenAndServe - Start app
func (a *App) ListenAndServe() {
	a.Init()
	a.Root.ListenAndServe(a.Config.Address)
}
