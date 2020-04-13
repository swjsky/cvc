package main

import (
	"encoding/json"
	"loreal.com/dit/utils"
	"log"
	"net/http"
)

/* 以下为具体 Endpoint 实现代码 */

//entryHandler - entry point for frontend web pages, to get initData in cookie
//endpoint: entry.html
//method: GET
func (a *App) entryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		showError(w, r, "", "无效的方法")
		return
	}
	q := r.URL.Query()
	// var userid int64
	// userid = -1
	appid := sanitizePolicy.Sanitize(q.Get("appid"))
	env := a.getEnv(appid)
	rt := a.getRuntime(env)
	// states := parseState(sanitizePolicy.Sanitize(q.Get("state")))
	scene := sanitizePolicy.Sanitize(q.Get("state"))
	_ = sanitizePolicy.Sanitize(q.Get("token"))

	openid := sanitizePolicy.Sanitize(q.Get("openid"))

	dataObject := map[string]interface{}{
		"appid":  appid,
		"scene":  scene,
		"openid": openid,
	}
	// follower, nickname := a.wxUserKeyInfo(openid)
	// dataObject["nickname"] = nickname
	if rt != nil {
		// if err := a.recordUser(
		// 	rt,
		// 	openid,
		// 	scene,
		// 	"0",
		// 	&userid,
		// ); err != nil {
		// 	log.Println("[ERR] - [EP][entry.html], err:", err)
		// }
	}
	if q.Get("debug") == "1" {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", " ")
		if err := encoder.Encode(dataObject); err != nil {
			log.Println("[ERR] - JSON encode error:", err)
			http.Error(w, "500", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	// cookieValue := url.PathEscape(utils.MarshalJSON(dataObject))
	if DEBUG {
		log.Println("[DEBUG] - set-cookie:", utils.MarshalJSON(dataObject))
	}
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "initdata",
	// 	Value:    cookieValue,
	// 	HttpOnly: false,
	// 	Secure:   false,
	// 	MaxAge:   0,
	// })
	http.ServeFile(w, r, "./public/index.html")
}
