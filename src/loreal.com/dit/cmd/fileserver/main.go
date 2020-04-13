//Package fileserver - An file server application
//It provice a basic web server for sharing files
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/module/modules/account"
	"loreal.com/dit/module/modules/resource"
	"loreal.com/dit/module/modules/root"
	"loreal.com/dit/utils"
)

const serviceName = "fileserver"
const serviceDescription = "ceh-File-Server"

//Cfg - config
var Cfg = &Config{
	Auth:           true,
	Address:        ":1502",
	Prefix:         "/",
	ShareFolder:    "./share",
	MongoConnStr:   "localhost",
	RedisServerStr: "localhost:6379",
}

func main() {

	utils.LoadOrCreateJSON("./config/config.json", &Cfg)

	flag.BoolVar(&Cfg.Auth, "auth", Cfg.Auth, "Enable Auth?")
	flag.StringVar(&Cfg.Address, "addr", Cfg.Address, "host:port of the LP service")
	flag.StringVar(&Cfg.Prefix, "prefix", Cfg.Prefix, "/path/ prefixe to LP service")
	flag.StringVar(&Cfg.ShareFolder, "share-folder", Cfg.ShareFolder, "path to share folder")
	flag.StringVar(&Cfg.RedisServerStr, "redis-server", Cfg.RedisServerStr, "redis server address")
	flag.StringVar(&Cfg.MongoConnStr, "mongo", Cfg.MongoConnStr, "MongoDB connection string")
	flag.Parse()
	// Prefix must be in the form of /path/
	Cfg.fixPrefix()

	endpoint.SetPrometheus("fileserver")

	//Create Main service
	rootModule := root.NewModule(serviceName, serviceDescription, Cfg.Prefix)
	uas := account.NewModule("account",
		Cfg.RedisServerStr, /*Redis server address*/
		3,                  /*Numbe of faild logins to lock the account */
		60*time.Second,     /*How long the account will stay locked*/
		7200*time.Second,   /*How long the token will be valid*/
	)

	registerHandlers(rootModule)
	registerEndpoints(rootModule, uas)

	rootModule.Install(
		uas,
		resource.NewModule(
			"resource",       //path
			Cfg.MongoConnStr, //mongo connection string
			"ev-resource",    //mongo db name
			"./share",        //upload path
			uas,              //user store
		),
	)

	rootModule.Start()
	go signalListen(rootModule)

	// defer func(m *root.Module) {
	// 	m.Dispose()
	// }(rootModule)

	log.Printf("CEH Service [%s]: Listening on %s serving %s", serviceName, Cfg.Address, Cfg.Prefix)
	if err := http.ListenAndServe(Cfg.Address, nil); err != nil {
		log.Fatal(err)
	}

}

func signalListen(root *root.Module) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	for {
		<-c
		root.Shutdown()
	}
}
