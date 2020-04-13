package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"loreal.com/dit/cmd/coupon-service/base"
	"loreal.com/dit/cmd/coupon-service/coupon"
	"loreal.com/dit/cmd/coupon-service/oauth"
	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"

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
		"api/events/":            {Handler: longPollingHandler},
		"api/coupontypes":        {Handler: couponTypeHandler},
		"api/coupons/":           {Handler: couponHandler},
		"api/redemptions":        {Handler: redemptionHandler},
		"api/apitester":          {Handler: apitesterHandler},
		// 管理员使用的redemption接口，用于紧急的数据修复
		"api/admin-redemptions": {Handler: redemptionMaintenanceHandler},
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
		// middlewares.WebTokenAuth(a.WebTokenAuthProvider),
		oauth.OauthCheckToken(),
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
	if appid == "ccs" {
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

func apitesterHandler(w http.ResponseWriter, r *http.Request) {
	switch apitest {
	case 1:
		{ //激活apitest功能
			coupon.ActivateTestedCoupontypes()
			apitest = 2
			w.WriteHeader(http.StatusOK)
			return
		}
	case 2:
		{ //已经激活了
			w.WriteHeader(http.StatusOK)
			return
		}
	default:
		{
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

//redemptionHandler - 关于兑换接口的入口
//endpoint: /api/redemption/
func redemptionHandler(w http.ResponseWriter, r *http.Request) {
	urlNodes := base.TrimURIPrefix(r.URL.RequestURI(), "")
	l := len(urlNodes)
	var lastNode = urlNodes[l-1]

	// token := r.Header.Get("Authorization")
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	mapclaims := base.GetMapClaim(token)
	if mapclaims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var requester = mapclaims.(*base.Requester)

	switch r.Method {
	case "POST":
		{
			switch lastNode {
			case "redemptions":
				{
					postCouponRedemption(requester, w, r)
				}
			default:
				{
					w.WriteHeader(http.StatusNotImplemented)
				}
			}

		}
	default:
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

// redemptionMaintenanceHandler - 维护时紧急数据修复，需要手动修复时使用此接口，不对客户端接入方公布。
func redemptionMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	urlNodes := base.TrimURIPrefix(r.URL.RequestURI(), "")
	l := len(urlNodes)
	var lastNode = urlNodes[l-1]

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	mapclaims := base.GetMapClaim(token)
	if mapclaims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var requester = mapclaims.(*base.Requester)

	switch r.Method {
	case "POST":
		{
			switch lastNode {
			case "admin-redemptions":
				{
					postMaintenanceCouponRedemption(requester, w, r)
				}
			default:
				{
					w.WriteHeader(http.StatusNotImplemented)
				}
			}
		}
	default:
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

//longPollingHandler
func longPollingHandler(w http.ResponseWriter, r *http.Request) {

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	mapclaims := base.GetMapClaim(token)
	if mapclaims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var requester = mapclaims.(*base.Requester)

	switch r.Method {
	case "GET":
		{
			msg, err := coupon.GetLatestCouponMessage(requester)
			if nil != err {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			outputJSON(w, &msg)
		}
	default:
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

//couponTypeHandler - 关于coupon 类型接口的入口
//endpoint: /api/coupontypes/
func couponTypeHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	mapclaims := base.GetMapClaim(token)
	if mapclaims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var requester = mapclaims.(*base.Requester)

	switch r.Method {
	case "GET":
		{
			getCouponTypes(requester, w, r)
		}
	default:
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func getCouponTypes(requester *base.Requester, w http.ResponseWriter, r *http.Request) {
	// consumerID := sanitizePolicy.Sanitize(r.Header.Get("consumerID"))
	// consumerRefID := sanitizePolicy.Sanitize(r.PostFormValue("consumerRefID"))
	// channelID := sanitizePolicy.Sanitize(r.PostFormValue("channelID"))

	ts, err := coupon.GetCouponTypes(requester)
	if nil != err {
		if sql.ErrConnDone == err {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if &(coupon.ErrRequesterForbidden) == err {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
		return
	}
	if nil == ts {
		ts = make([]*coupon.PublishedCouponType, 0)
	}
	outputJSON(w, ts)
}

//couponHandler - 关于coupon 接口的入口
//endpoint: /api/coupons/
func couponHandler(w http.ResponseWriter, r *http.Request) {
	urlNodes := base.TrimURIPrefix(r.URL.RequestURI(), "")
	l := len(urlNodes)
	var lastNode = urlNodes[l-1]

	// token := r.Header.Get("Authorization")
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	mapclaims := base.GetMapClaim(token)
	if mapclaims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var requester = mapclaims.(*base.Requester)

	switch r.Method {
	case "POST":
		{
			switch lastNode {
			case "coupons":
				{
					postCoupon(requester, w, r)
				}
			default:
				{
					w.WriteHeader(http.StatusNotImplemented)
				}
			}

		}
	case "GET":
		{
			switch lastNode {
			case "coupons":
				{
					getCoupons(requester, w, r)
				}
			default:
				{
					getCoupon(requester, lastNode, w, r)
				}
			}
		}
	default:
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func getCoupons(requester *base.Requester, w http.ResponseWriter, r *http.Request) {
	consumerID := sanitizePolicy.Sanitize(r.Header.Get("consumerID"))
	transType := sanitizePolicy.Sanitize(r.Header.Get("transType"))

	var err error
	var cs []*coupon.Coupon
	if base.IsBlankString(transType) {
		cs, err = coupon.GetCoupons(requester, consumerID, "")
	} else {
		cs, err = coupon.GetCouponsWithTransactions(requester, consumerID, "", transType)
	}

	for _, c := range cs {
		expired, _ := coupon.ValidateCouponExpired(requester, c)
		if expired {
			c.State = coupon.SExpired
		}
	}

	if nil != err {
		if sql.ErrConnDone == err {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if &(coupon.ErrRequesterForbidden) == err {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
		return
	}
	if nil == cs {
		cs = make([]*coupon.Coupon, 0)
	}
	outputJSON(w, cs)
}

func getCoupon(requester *base.Requester, couponID string, w http.ResponseWriter, r *http.Request) {
	if !requester.HasRole(base.ROLE_COUPON_ISSUER) && !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	transType := sanitizePolicy.Sanitize(r.Header.Get("transType"))
	var err error
	var c *coupon.Coupon

	if base.IsBlankString(transType) {
		c, err = coupon.GetCoupon(requester, sanitizePolicy.Sanitize(couponID))
	} else {
		c, err = coupon.GetCouponWithTransactions(requester, sanitizePolicy.Sanitize(couponID), transType)
	}

	if nil != err {
		if sql.ErrConnDone == err {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if &(coupon.ErrRequesterForbidden) == err {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
		return
	}
	if nil == c {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	expired, err := coupon.ValidateCouponExpired(requester, c)

	if nil != err {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if expired {
		c.State = coupon.SExpired
	}

	outputJSON(w, &c)
}

func postCoupon(requester *base.Requester, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("[ERR] - [gatewayHandler][ParseForm], err: %v", err)
	}
	// 当consumerID有值时，表示为单个消费者产生卡券，
	// 如果没有consumerID或者consumerID没有值，那么去判断consumerIDs
	// consumerIDs是批量创建卡券
	consumerIDs := sanitizePolicy.Sanitize(r.PostFormValue("consumerIDs"))
	consumerRefIDs := sanitizePolicy.Sanitize(r.PostFormValue("consumerRefIDs"))
	couponTypeID := sanitizePolicy.Sanitize(r.PostFormValue("couponTypeID"))
	channelID := sanitizePolicy.Sanitize(r.PostFormValue("channelID"))

	if !base.IsValidUUID(couponTypeID) {
		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", &coupon.ErrCouponTemplateNotFound)
		return
	}

	if base.IsBlankString(consumerIDs) {
		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", &coupon.ErrConsumerIDInvalid)
		return
	}

	// 批量模式签发卡券
	cs, errMap, err := coupon.IssueCoupons(requester, consumerIDs, consumerRefIDs, channelID, couponTypeID)
	if nil != err {
		if &(coupon.ErrRequesterForbidden) == err {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
		return
	}

	if nil != errMap && len(errMap) > 0 {
		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", errMap)
		return
	}

	outputJSON(w, cs)
}

// couponID 和 couponTypeID 只能二选一。
// 当 couponID 不为空时，根据 couponID 来核销。
// 当 couponID 为空时，并且 couponTypeID 不为空，则根据 couponTypeID 来核销，但要判断是否只有一张卡券，如果有多张，将会返回错误。
func postCouponRedemption(requester *base.Requester, w http.ResponseWriter, r *http.Request) {
	if !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("[ERR] - [gatewayHandler][ParseForm], err: %v", err)
	}

	couponID := sanitizePolicy.Sanitize(r.PostFormValue("couponID"))

	consumerIDs := sanitizePolicy.Sanitize(r.PostFormValue("consumerIDs"))
	couponTypeID := sanitizePolicy.Sanitize(r.PostFormValue("couponTypeID"))

	// extraInfo := sanitizePolicy.Sanitize(r.PostFormValue("extraInfo"))
	extraInfo := r.PostFormValue("extraInfo")

	if base.IsBlankString(consumerIDs) {
		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", &coupon.ErrConsumerIDInvalid)
		return
	}

	if base.IsBlankString(couponID) && !base.IsBlankString(couponTypeID) && base.IsValidUUID(couponTypeID) {
		//根据卡券类型核销唯一的一张卡券
		errMap, err := coupon.RedeemCouponByType(requester, consumerIDs, couponTypeID, extraInfo)
		if nil != err {
			if &(coupon.ErrRequesterForbidden) == err {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
			return
		}

		if nil != errMap && len(errMap) > 0 {
			base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", errMap)
			return
		}
		return
	}

	// 根据卡券ID核销唯一的一张卡券
	if base.IsBlankString(couponID) || !base.IsValidUUID(couponID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	errs, err := coupon.RedeemCoupon(requester, consumerIDs, couponID, extraInfo)
	if nil != err {
		if &(coupon.ErrRequesterForbidden) == err {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
		return
	}

	if nil != errs && len(errs) > 0 {
		base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", errs[0])
		return
	}
}

// postMaintenanceCouponRedemption - 临时数据修复接口， 不要对接入客户端暴露。暂时只支持通过couponIDs强行核销问题卡券。
func postMaintenanceCouponRedemption(requester *base.Requester, w http.ResponseWriter, r *http.Request) {
	if !requester.HasRole(base.ROLE_COUPON_REDEEMER) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("[ERR] - [gatewayHandler][ParseForm], err: %v", err)
	}

	couponIDs := sanitizePolicy.Sanitize(r.PostFormValue("couponIDs"))
	extraInfo := r.PostFormValue("extraInfo")

	err := coupon.RedeemCouponsInMaintenance(requester, couponIDs, extraInfo)

	if err != nil {
		fmt.Println(err)
	}

}

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
