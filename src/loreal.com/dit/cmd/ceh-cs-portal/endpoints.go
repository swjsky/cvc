package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"

	"loreal.com/dit/cmd/ceh-cs-portal/restful"

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

func brandFilter(r *http.Request, item *map[string]interface{}) bool {
	roles := r.Header.Get("roles")
	if roles == "admin" {
		return false
	}
	loginBrand := r.Header.Get("brand")
	if loginBrand == "" {
		return false
	}
	targetBrand, _ := ((*item)["brand"]).(*string)
	return *targetBrand != "" && *targetBrand != loginBrand
}

func (a *App) initEndpoints() {
	rt := a.getRuntime("prod")
	a.Endpoints = map[string]EndpointEntry{
		"api/kvstore":            {Handler: a.kvstoreHandler, Middlewares: a.noAuthMiddlewares("api/kvstore")},
		"api/visit":              {Handler: a.pvHandler},
		"error":                  {Handler: a.errorHandler, Middlewares: a.noAuthMiddlewares("error")},
		"debug":                  {Handler: a.debugHandler},
		"maintenance/fe/upgrade": {Handler: a.feUpgradeHandler},
		"api/gw":                 {Handler: a.gatewayHandler},
		"api/brand/": {
			Handler: restful.NewHandler(
				"brand",
				restful.NewSQLiteAdapter(rt.db,
					rt.mutex,
					"Brand",
					Brand{},
				),
			).ServeHTTP,
		},
		// "api/customer/": {
		// 	Handler: restful.NewHandler(
		// 		"customer",
		// 		restful.NewSQLiteAdapter(rt.db,
		// 			rt.mutex,
		// 			"Customer",
		// 			Customer{},
		// 		),
		// 	).SetFilter(storeFilter).ServeHTTP,
		// },
	}

	postPrepareDB(rt)
}

//noAuthMiddlewares - middlewares without auth
func (a *App) noAuthMiddlewares(path string) []endpoint.ServerMiddleware {
	return []endpoint.ServerMiddleware{
		middlewares.NoCache(),
		middlewares.ServerInstrumentation(path, endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	}
}

// //webTokenAuthMiddlewares - middlewares auth by token
// func (a *App) webTokenAuthMiddlewares(path string) []endpoint.ServerMiddleware {
// 	return []endpoint.ServerMiddleware{
// 		middlewares.NoCache(),
// 		middlewares.WebTokenAuth(a.WebTokenAuthProvider),
// 		middlewares.ServerInstrumentation(path, endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
// 	}
// }

//getDefaultMiddlewares - middlewares installed by defaults
func (a *App) getDefaultMiddlewares(path string) []endpoint.ServerMiddleware {
	return []endpoint.ServerMiddleware{
		middlewares.NoCache(),
		middlewares.WebTokenAuth(a.WebTokenAuthProvider),
		// middlewares.BasicAuthOrTokenAuthWithRole(a.AuthProvider, "", "user,admin"),
		middlewares.ServerInstrumentation(
			path,
			endpoint.RequestCounter,
			endpoint.LatencyHistogram,
			endpoint.DurationsSummary,
		),
	}
}

func (a *App) getEnv(appid string) string {
	if appid == "" {
		if a.Config.Production {
			return "prod"
		}
		return "pp"
	}
	if appid == "ceh" {
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
		outputJSON(w, APIStatus{
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
		outputJSON(w, APIStatus{
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
		outputJSON(w, APIStatus{
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
	rt := a.getRuntime(r.PostForm.Get("env"))
	if rt == nil {
		outputJSON(w, map[string]interface{}{
			"code": -2,
			"msg":  "Invalid APPID",
		})
		return
	}
	userid, _ := strconv.ParseInt(sanitizePolicy.Sanitize(r.PostForm.Get("userid")), 10, 64)
	pageid := sanitizePolicy.Sanitize(q.Get("pageid"))
	scene := sanitizePolicy.Sanitize(q.Get("scene"))
	visitState, _ := strconv.Atoi(sanitizePolicy.Sanitize(q.Get("type")))

	if err := a.recordPV(
		rt,
		userid,
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

//postPrepareDB - initialized database after init endpoints
func postPrepareDB(rt *RuntimeEnv) {
	//init database tables
	sqlStmts := []string{
		// 		`CREATE TRIGGER IF NOT EXISTS insert_fulfill INSERT ON fulfillment
		// BEGIN
		//     UPDATE CustomerOrder SET qtyfulfilled=qtyfulfilled+new.quantity WHERE id=new.orderid;
		// END;`,
		// 		`CREATE TRIGGER IF NOT EXISTS delete_fulfill DELETE ON fulfillment
		// BEGIN
		//     UPDATE CustomerOrder SET qtyfulfilled=qtyfulfilled-old.quantity WHERE id=old.orderid;
		// END;`,
		// 		`CREATE TRIGGER IF NOT EXISTS before_update_fulfill BEFORE UPDATE ON fulfillment
		// BEGIN
		//     UPDATE CustomerOrder SET qtyfulfilled=qtyfulfilled-old.quantity WHERE id=old.orderid;
		// END;`,
		// 		`CREATE TRIGGER IF NOT EXISTS after_update_fulfill AFTER UPDATE ON fulfillment
		// BEGIN
		//     UPDATE CustomerOrder SET qtyfulfilled=qtyfulfilled+new.quantity WHERE id=new.orderid;
		// END;`,
		// 		"CREATE UNIQUE INDEX IF NOT EXISTS uidxOpenID ON WxUser(OpenID);",
	}

	log.Printf("[INFO] - Post Prepare DB for [%s]...\n", rt.Config.Name)
	for _, sqlStmt := range sqlStmts {
		_, err := rt.db.Exec(sqlStmt)
		if err != nil {
			log.Printf("[ERR] - [PrepareDB] %q: %s\n", err, sqlStmt)
			return
		}
	}
	rt.stmts = make(map[string]*sql.Stmt, 0)
	log.Printf("[INFO] - DB for [%s] prepared!\n", rt.Config.Name)
}
