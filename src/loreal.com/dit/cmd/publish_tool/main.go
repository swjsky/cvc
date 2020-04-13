//Package publish_tool - An CEH application
//LP self registration
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/module/modules/root"
	"loreal.com/dit/utils"
)

const serviceName = "MemberCenter"
const serviceDescription = "MemberCenter"

//Cfg - config
var Cfg = &Config{
	Address:  ":1507",
	Prefix:   "/",
	BasePath: "E:\\Go\\src\\loreal.com\\dit\\cmd",
}

func main() {
	utils.LoadOrCreateJSON("./config/config.json", &Cfg)

	flag.StringVar(&Cfg.Address, "addr", Cfg.Address, "host:port of the LP service")
	flag.StringVar(&Cfg.Prefix, "prefix", Cfg.Prefix, "/path/ prefixe to LP service")
	flag.StringVar(&Cfg.BasePath, "base path", Cfg.BasePath, "Loreal projects path")
	flag.Parse()
	// Prefix must be in the form of /path/
	Cfg.fixPrefix()

	endpoint.SetPrometheus("MemberCenter")

	//Create Main service
	rootModule := root.NewModule(serviceName, serviceDescription, Cfg.Prefix)

	rootModule.Install(
		NewModule(""),
	)

	rootModule.Start()
	go signalListen(rootModule)

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
