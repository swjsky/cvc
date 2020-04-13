package main

import (
	"loreal.com/dit/endpoint"
	"loreal.com/dit/loreal/webservice"
	"loreal.com/dit/middlewares"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/microcosm-cc/bluemonday"
)

// var seededRand *rand.Rand
var sanitizePolicy *bluemonday.Policy

var errorTemplate *template.Template

func init() {
	// 	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	sanitizePolicy = bluemonday.UGCPolicy()

	var err error
	errorTemplate, _ = template.ParseFiles("./template/error.tpl")
	if err != nil {
		log.Panic("[ERR] - Parsing error template", err)
	}
}

func (a *App) initEndpoints() {
	a.Endpoints = map[string]EndpointEntry{
		"entry.html":   {Handler: a.entryHandler, Middlewares: a.noAuthMiddlewares("entry.html")},
		"api/kvstore":  {Handler: a.kvstoreHandler, Middlewares: a.noAuthMiddlewares("api/kvstore")},
		"api/visit":    {Handler: a.pvHandler, Middlewares: a.noAuthMiddlewares("api/visit")},
		"error":        {Handler: a.errorHandler, Middlewares: a.noAuthMiddlewares("error")},
		"report/visit": {Handler: a.reportVisitHandler},
	}
}

//noAuthMiddlewares - middlewares without auth
func (a *App) noAuthMiddlewares(path string) []endpoint.ServerMiddleware {
	return []endpoint.ServerMiddleware{
		middlewares.NoCache(),
		middlewares.ServerInstrumentation(path, endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	}
}

//tokenAuthMiddlewares - middlewares auth by token
// func (a *App) tokenAuthMiddlewares(path string) []endpoint.ServerMiddleware {
// 	return []endpoint.ServerMiddleware{
// 		middlewares.NoCache(),
// 		middlewares.ServerInstrumentation(path, endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
// 		a.signatureVerifier(),
// 	}
// }

//getDefaultMiddlewares - middlewares installed by defaults
func (a *App) getDefaultMiddlewares(path string) []endpoint.ServerMiddleware {
	return []endpoint.ServerMiddleware{
		middlewares.NoCache(),
		middlewares.BasicAuthOrTokenAuthWithRole(a.AuthProvider, "", "user,admin"),
		middlewares.ServerInstrumentation(
			path,
			endpoint.RequestCounter,
			endpoint.LatencyHistogram,
			endpoint.DurationsSummary,
		),
	}
}

func (a *App) getEnv(appid string) string {
	if appid == a.Config.AppID {
		return "prod"
	}
	return "pp"
}

/* 以下为具体 Endpoint 实现代码 */

//errorHandler - query error info
//endpoint: error
//method: GET
func (a *App) errorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	q := r.URL.Query()
	title := sanitizePolicy.Sanitize(q.Get("title"))
	errmsg := sanitizePolicy.Sanitize(q.Get("errmsg"))

	if err := errorTemplate.Execute(w, map[string]interface{}{
		"title":  title,
		"errmsg": errmsg,
	}); err != nil {
		log.Println("[ERR] - errorTemplate error:", err)
		http.Error(w, "500", http.StatusInternalServerError)
	}
}

/* 以下为具体 Endpoint 实现代码 */

//kvstoreHandler - get value from kvstore in runtime
//endpoint: /api/kvstore
//method: GET
func (a *App) kvstoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		outputJSON(w, webservice.APIStatus{
			ErrCode:    -100,
			ErrMessage: "Method not acceptable",
		})
		return
	}
	q := r.URL.Query()
	ticket := q.Get("ticket")
	env := a.getEnv(q.Get("appid"))
	rt := a.getRuntime(env)
	if rt == nil {
		outputJSON(w, webservice.APIStatus{
			ErrCode:    -1,
			ErrMessage: "invalid appid",
		})
		return
	}
	var result struct {
		Value interface{} `json:"value"`
	}
	var ok bool
	var v interface{}
	v, ok = rt.Retrive(ticket)
	if !ok {
		outputJSON(w, webservice.APIStatus{
			ErrCode:    -2,
			ErrMessage: "invalid ticket",
		})
		return
	}
	switch val := v.(type) {
	case chan interface{}:
		// log.Println("[Hu Bin] - Get Value Chan:", val)
		result.Value = <-val
		// log.Println("[Hu Bin] - Get Value from Chan:", result.Value)
	default:
		// log.Println("[Hu Bin] - Get Value:", val)
		result.Value = val
	}
	outputJSON(w, result)
}

//pvHandler - record PV/UV
//endpoint: /api/visit
//method: GET
func (a *App) pvHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		outputJSON(w, map[string]interface{}{
			"code": -1,
			"msg":  "Not support",
		})
		return
	}
	q := r.URL.Query()
	appid := q.Get("appid")
	env := a.getEnv(appid)
	rt := a.getRuntime(env)
	if rt == nil {
		log.Println("[ERR] - Invalid appid:", appid)
		outputJSON(w, map[string]interface{}{
			"code": -2,
			"msg":  "Invalid APPID",
		})
		return
	}
	openid := sanitizePolicy.Sanitize(q.Get("openid"))
	pageid := sanitizePolicy.Sanitize(q.Get("pageid"))
	scene := sanitizePolicy.Sanitize(q.Get("scene"))
	visitState, _ := strconv.Atoi(sanitizePolicy.Sanitize(q.Get("type")))

	if err := a.recordPV(
		rt,
		openid,
		pageid,
		scene,
		visitState,
	); err != nil {
		log.Println("[ERR] - [EP][api/visit], err:", err)
		outputJSON(w, map[string]interface{}{
			"code": -3,
			"msg":  "internal error",
		})
		return
	}
	outputJSON(w, map[string]interface{}{
		"code": 0,
		"msg":  "ok",
	})
}

//CSV BOM
//file.Write([]byte{0xef, 0xbb, 0xbf})

func outputExcel(w http.ResponseWriter, b []byte, filename string) {
	w.Header().Add("Content-Disposition", "attachment; filename="+filename)
	//w.Header().Add("Content-Type", "application/vnd.ms-excel")
	w.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	//	w.Header().Add("Content-Transfer-Encoding", "binary")
	w.Write(b)
}

func outputText(w http.ResponseWriter, b []byte) {
	w.Header().Add("Content-Type", "text/plain;charset=utf-8")
	w.Write(b)
}

func showError(w http.ResponseWriter, r *http.Request, title, message string) {
	if err := errorTemplate.Execute(w, map[string]interface{}{
		"title":  title,
		"errmsg": message,
	}); err != nil {
		log.Println("[ERR] - errorTemplate error:", err)
		http.Error(w, "500", http.StatusInternalServerError)
	}
}
