//Smartfix
package main

import (
	"encoding/json"
	// "fmt"
	"flag"
	"log"
	"math/rand"
	"time"

	"loreal.com/dit/cmd/coupon-service/base"
	"loreal.com/dit/cmd/coupon-service/coupon"
	"loreal.com/dit/utils"

	"github.com/gobuffalo/packr/v2"
)

// TODO: 重构离散的变量
// TODO: 各种规则自己注册到引擎
// TODO: Coupon的规则应用hash值
// TODO: api请求带上签名
// TODO: 模板可以修改并应用到后申请的卡券
// TODO: 模板可以撤回
// TODO: 规则可以被禁用启用，可以影响到已经发行的卡券

//Version - generate on build time by makefile
var Version = "v0.1"

//CommitID - generate on build time by makefile
var CommitID = ""

const serviceName = "Coupon-Service"
const serviceDescription = "A service to issue/redeem coupons."

var apitest int = 0
var app *App

func init() {

}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("[INFO] -", serviceName, Version+"-"+CommitID)
	log.Println("[INFO] -", serviceDescription)

	utils.LoadOrCreateJSON("./config/config.json", &base.Cfg) //cfg initialized in config.go

	var preDefinedData preDefinedDataForDatabase
	box := packr.New("pre-defined", "./pre-defined")
	jsonString, err := box.FindString("predefined-data.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(jsonString), &preDefinedData)
	if err != nil {
		panic(err)
	}
	// utils.LoadOrCreateJSON("./config/predefined-data.json", &preDefinedData)

	flag.StringVar(&base.Cfg.Address, "addr", base.Cfg.Address, "host:port of the service")
	flag.StringVar(&base.Cfg.Prefix, "prefix", base.Cfg.Prefix, "/path/ prefixe to service")
	flag.IntVar(&apitest, "apitest", 0, "")
	flag.Parse()

	//Create Main service
	var app = NewApp(serviceName, serviceDescription, &base.Cfg)
	app.Init()
	var rt = app.getRuntime("prod")
	for _, t := range preDefinedData.PublishedCouponTypes {
		t.InitRules()
	}
	coupon.Init(preDefinedData.CouponTemplates, preDefinedData.PublishedCouponTypes, preDefinedData.SupportedRules, rt.db, base.Cfg.JwtKey)
	app.ListenAndServe()
}
