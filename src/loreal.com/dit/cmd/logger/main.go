package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"loreal.com/dit/cmd/logger/modules/websocket"
	"loreal.com/dit/module/modules/account"
	"loreal.com/dit/utils"

	mgo "gopkg.in/mgo.v2"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

const (
	serviceName        = "Logger Service"
	serviceDescription = "CEH RPC Logger service"
	loggerCollName     = "logger"
)

func main() {

	utils.LoadOrCreateJSON("./config/config.json", &cfg) //cfg initialized in config.go
	flag.StringVar(&cfg.Address, "addr", cfg.Address, "host:port of NPS service")
	flag.StringVar(&cfg.RPCAddress, "rpc_addr", cfg.RPCAddress, "host:port of NPS service")
	flag.StringVar(&cfg.Prefix, "prefix", cfg.Prefix, "/path/ prefixe to NPS service")
	flag.StringVar(&cfg.RedisServerStr, "redis", cfg.RedisServerStr, "Redis connection string")
	flag.Parse()

	//Create Main service
	app := NewApp(serviceName, serviceDescription, &cfg)
	uas := account.NewModule("account",
		app.Config.RedisServerStr, /*Redis server address*/
		3,                         /*Numbe of faild logins to lock the account */
		60*time.Second,            /*How long the account will stay locked*/
		7200*time.Second,          /*How long the token will be valid*/
	)
	app.Root.Install(
		uas,
		websocket.NewModule("websocket"),
	)
	go signalListen(app)
	app.AuthProvider = uas
	app.InitDB(cfg.SqliteDBName)
	if err := app.initMongoIndexes(); err != nil {
		panic(err)
	}
	// start rpc server
	go app.Start()
	app.ListenAndServe()
}

func (a *App) initMongoIndexes() (err error) {
	session, err := a.MgoSessionManager.Get()
	if err != nil {
		return
	}
	defer session.Close()
	coll := session.DB(a.Config.MongoDBName).C(loggerCollName)
	err = coll.EnsureIndex(mgo.Index{Key: []string{"project", "created_at"}, Unique: false})
	if err != nil {
		return
	}
	return
}

func signalListen(a *App) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	for {
		<-c
		a.Root.Shutdown()
	}
}
