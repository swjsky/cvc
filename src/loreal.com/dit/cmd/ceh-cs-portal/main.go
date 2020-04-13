//Smartfix
package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"loreal.com/dit/module/modules/account"
	"loreal.com/dit/utils"
)

//Version - generate on build time by makefile
var Version = "v0.1"

//CommitID - generate on build time by makefile
var CommitID = ""

const serviceName = "CS-PORTAL"
const serviceDescription = "CEH-CS-PORTAL"

var app *App

func init() {

}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("[INFO] -", serviceName, Version+"-"+CommitID)
	log.Println("[INFO] -", serviceDescription)

	utils.LoadOrCreateJSON("./config/config.json", &cfg) //cfg initialized in config.go

	flag.StringVar(&cfg.Address, "addr", cfg.Address, "host:port of the service")
	flag.StringVar(&cfg.Prefix, "prefix", cfg.Prefix, "/path/ prefixe to service")
	flag.Parse()

	//Create Main service
	var app = NewApp(serviceName, serviceDescription, &cfg)
	uas := account.NewModule("account",
		serviceName,        /*Token Issuer*/
		[]byte(cfg.JwtKey), /*Json Web Token Sign Key*/
		10,                 /*Numbe of faild logins to lock the account */
		60*time.Second,     /*How long the account will stay locked*/
		7200*time.Second,   /*How long the token will be valid*/
	)
	app.Root.Install(
		uas,
	)
	app.AuthProvider = uas
	app.WebTokenAuthProvider = uas
	app.ListenAndServe()
}
