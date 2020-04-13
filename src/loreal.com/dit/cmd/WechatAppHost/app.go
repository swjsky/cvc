package main

import (
	"database/sql"
	"loreal.com/dit/module"
	"loreal.com/dit/module/modules/root"
	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
	"loreal.com/dit/utils"
	"loreal.com/dit/utils/task"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/robfig/cron"
)

//App - data struct for App & configuration file
type App struct {
	Name            string
	Description     string
	Config          *Config
	Root            *root.Module
	Endpoints       map[string]EndpointEntry
	MessageHandlers map[string]func(*module.Message) bool
	AuthProvider    middlewares.RoleVerifier
	Scheduler       *cron.Cron
	TaskManager     *task.Manager
	wg              *sync.WaitGroup
	mutex           *sync.RWMutex
	Runtime         map[string]*RuntimeEnv
}

//RuntimeEnv - runtime env
type RuntimeEnv struct {
	Config  *Env
	stmts   map[string]*sql.Stmt
	db      *sql.DB
	KVStore map[string]interface{}
	mutex   *sync.RWMutex
}

//Get - get value from kvstore in memory
func (rt *RuntimeEnv) Get(key string) (value interface{}, ok bool) {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()
	value, ok = rt.KVStore[key]
	return
}

//Retrive - get value from kvstore in memory, and delete it
func (rt *RuntimeEnv) Retrive(key string) (value interface{}, ok bool) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	value, ok = rt.KVStore[key]
	if ok {
		delete(rt.KVStore, key)
	}
	return
}

//Set - set value to kvstore in memory
func (rt *RuntimeEnv) Set(key string, value interface{}) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	rt.KVStore[key] = value
}

//EndpointEntry - endpoint registry entry
type EndpointEntry struct {
	Handler     func(http.ResponseWriter, *http.Request)
	Middlewares []endpoint.ServerMiddleware
}

//NewApp - create new app
func NewApp(name, description string, config *Config) *App {
	if config == nil {
		log.Println("Missing configuration data")
		return nil
	}
	endpoint.SetPrometheus(strings.Replace(name, "-", "_", -1))
	app := &App{
		Name:            name,
		Description:     description,
		Config:          config,
		Root:            root.NewModule(name, description, config.Prefix),
		Endpoints:       make(map[string]EndpointEntry, 0),
		MessageHandlers: make(map[string]func(*module.Message) bool, 0),
		Scheduler:       cron.New(),
		wg:              &sync.WaitGroup{},
		mutex:           &sync.RWMutex{},
		Runtime:         make(map[string]*RuntimeEnv),
	}
	app.TaskManager = task.NewManager(app, 100)
	return app
}

//Init - app initialization
func (a *App) Init() {
	if a.Config != nil {
		a.Config.fixPrefix()
		for _, env := range a.Config.Envs {
			utils.MakeFolder(env.DataFolder)
			a.Runtime[env.Name] = &RuntimeEnv{
				Config:  env,
				KVStore: make(map[string]interface{}, 1024),
				mutex:   &sync.RWMutex{},
			}
		}
		a.InitDB()
	}
	a.registerEndpoints()
	a.registerMessageHandlers()
	a.registerTasks()
	// utils.LoadOrCreateJSON("./saved_status.json", &a.Status)
	a.Root.OnStop = func(p *module.Module) {
		a.TaskManager.SendAll("stop")
		a.wg.Wait()
	}
	a.Root.OnDispose = func(p *module.Module) {
		for _, env := range a.Runtime {
			if env.db != nil {
				log.Println("Close sqlite for", env.Config.Name)
				env.db.Close()
			}
		}
		// utils.SaveJSON(a.Status, "./saved_status.json")
	}
}

//registerEndpoints - Register Endpoints
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

//registerMessageHandlers - Register Message Handlers
func (a *App) registerMessageHandlers() {
	a.initMessageHandlers()
	for path, handler := range a.MessageHandlers {
		a.Root.AddMessageHandler(path, handler)
	}
}

//StartScheduler - register and start the scheduled tasks
func (a *App) StartScheduler() {
	if a.Scheduler == nil {
		a.Scheduler = cron.New()
	} else {
		a.Scheduler.Stop()
		a.Scheduler = cron.New()
	}
	for _, item := range a.Config.ScheduledTasks {
		log.Println("[INFO] - Adding task:", item.Task)
		func() {
			s := item.Schedule
			t := item.Task
			a.Scheduler.AddFunc(s, func() {
				a.TaskManager.RunTask(t, item.DefaultArgs...)
			})
		}()
	}
	a.Scheduler.Start()
}

//ListenAndServe - Start app
func (a *App) ListenAndServe() {
	a.Init()
	a.StartScheduler()
	a.Root.ListenAndServe(a.Config.Address)
}
