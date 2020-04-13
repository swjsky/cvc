//General Wechat WebAPP Host
package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"loreal.com/dit/module/modules/account"
	"loreal.com/dit/module/modules/wechat"
	"loreal.com/dit/utils"
)

//Version - generate on build time by makefile
var Version = "v0.1"

//CommitID - generate on build time by makefile
var CommitID = ""

func main() {
	rand.Seed(time.Now().UnixNano())
	const serviceName = "wxAppHost"
	const serviceDescription = "Wechat WebAPP Host"
	log.Println("[INFO] -", serviceName, Version+"-"+CommitID)
	log.Println("[INFO] -", serviceDescription)

	utils.LoadOrCreateJSON("./config/config.json", &cfg) //cfg initialized in config.go

	flag.StringVar(&cfg.Address, "addr", cfg.Address, "host:port of the service")
	flag.StringVar(&cfg.Prefix, "prefix", cfg.Prefix, "/path/ prefixe to service")
	flag.StringVar(&cfg.RedisServerStr, "redis", cfg.RedisServerStr, "Redis connection string")
	flag.StringVar(&cfg.AppDomainName, "app-domain", cfg.AppDomainName, "app domain name")
	flag.Parse()

	//Create Main service
	var app = NewApp(serviceName, serviceDescription, &cfg)
	uas := account.NewModule("account",
		app.Config.RedisServerStr, /*Redis server address*/
		3,                         /*Numbe of faild logins to lock the account */
		60*time.Second,            /*How long the account will stay locked*/
		7200*time.Second,          /*How long the token will be valid*/
	)
	app.Root.Install(
		uas,
		wechat.NewModuleWithCEHTokenService(
			"wx",
			app.Config.AppDomainName,
			app.Config.TokenServiceURL,
			app.Config.TokenServiceUsername,
			app.Config.TokenServicePassword,
		),
	)
	app.AuthProvider = uas
	app.ListenAndServe()
}
