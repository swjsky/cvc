package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//DEBUG - whether in debug mode
var DEBUG bool

//INFOLEVEL - info level for debug mode
var INFOLEVEL int

//LOGLEVEL - info level for logs
var LOGLEVEL int

func init() {
	if os.Getenv("EV_DEBUG") != "" {
		DEBUG = true
	}
	INFOLEVEL = 1
	LOGLEVEL = 1
}

func retry(count int, fn func() error) error {
	total := count
retry:
	err := fn()
	if err != nil {
		count--
		log.Println("[INFO] - Retry: ", total-count)
		if count > 0 {
			goto retry
		}
	}
	return err
}

func parseState(state string) map[string]string {
	result := make(map[string]string, 2)
	var err error
	state, err = url.PathUnescape(state)
	if err != nil {
		log.Println("[ERR] - parseState", err)
		return result
	}
	if DEBUG {
		log.Println("[DEBUG] - PathUnescape state:", state)
	}
	states := strings.Split(state, ";")
	for _, kv := range states {
		sp := strings.Index(kv, ":")
		if sp < 0 {
			//empty value
			result[kv] = ""
			continue
		}
		result[kv[:sp]] = kv[sp+1:]
	}
	return result
}

func (a *App) getRuntime(env string) *RuntimeEnv {
	runtime, ok := a.Runtime[env]
	if !ok {
		return nil
	}
	return runtime
}

//getStmt - get stmt from app safely
func (a *App) getStmt(runtime *RuntimeEnv, name string) *sql.Stmt {
	runtime.mutex.RLock()
	defer runtime.mutex.RUnlock()
	if stmt, ok := runtime.stmts[name]; ok {
		return stmt
	}
	return nil
}

//getStmt - get stmt from app safely
func (a *App) setStmt(runtime *RuntimeEnv, name, query string) (stmt *sql.Stmt, err error) {
	stmt, err = runtime.db.Prepare(query)
	if err != nil {
		logError(err, name)
		return nil, err
	}
	runtime.mutex.Lock()
	runtime.stmts[name] = stmt
	runtime.mutex.Unlock()
	return stmt, nil
}

//outputJSON - output json for http response
func outputJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	enc := json.NewEncoder(w)
	if DEBUG {
		enc.SetIndent("", " ")
	}
	if err := enc.Encode(data); err != nil {
		log.Println("[ERR] - [outputJSON] JSON encode error:", err)
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}
}

func logError(err error, msg string) {
	if err != nil {
		log.Printf("[ERR] - %s, err: %v\n", msg, err)
	}
}

func debugInfo(source, msg string, level int) {
	if DEBUG && INFOLEVEL >= level {
		log.Printf("[DEBUG] - [%s]%s\n", source, msg)
	}
}
