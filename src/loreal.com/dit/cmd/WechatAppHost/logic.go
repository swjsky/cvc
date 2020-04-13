package main

import (
	"database/sql"
	"encoding/json"
	"loreal.com/dit/module/modules/loreal"
	"loreal.com/dit/wechat"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//DEBUG - whether in debug mode
var DEBUG bool

//INFOLEVEL - info level for debug mode
var INFOLEVEL int

//LOGLEVEL - info level for logs
var LOGLEVEL int

var wxAccount *wechat.Account

func init() {
	wxAccount = wechat.Accounts["default"]
	if wxAccount.Token.Requester == nil {
		wxAccount.Token.Requester = loreal.NewWechatTokenService(wxAccount)
	}
	if os.Getenv("EV_DEBUG") != "" {
		DEBUG = true
	}
	INFOLEVEL = 1
	LOGLEVEL = 1
}

//var wxAccount = wechat.Accounts["default"]

func lorealCardValid(cardNo string) bool {
	if len(cardNo) != 22 {
		return false
	}
	re := regexp.MustCompile("\\d{22}")
	if !re.MatchString(cardNo) {
		return false
	}
	return (cardNo == lorealCardCheckSum(cardNo[:20]))
}

//lorealCardCheckSum calculate 2 check sum bit for loreal card
func lorealCardCheckSum(cardNo string) string {
	firstLineWeight := []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2}
	secondLineWeight := []int{4, 3, 2, 7, 6, 5, 4, 3, 2, 7, 6, 5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	cardRunes := []rune(cardNo)

	var firstDigital, secondDigital int

	var temp, result int
	for i := 0; i <= 19; i++ {
		cardBit, _ := strconv.Atoi(string(cardRunes[i]))
		temp = cardBit * firstLineWeight[i]
		if temp > 9 {
			temp -= 9
		}
		result += temp
	}

	firstDigital = result % 10
	if firstDigital != 0 {
		firstDigital = 10 - firstDigital
	}

	cardNo = fmt.Sprintf("%s%d", cardNo, firstDigital)
	cardRunes = []rune(cardNo)

	result = 0

	for i := 0; i <= 20; i++ {
		cardBit, _ := strconv.Atoi(string(cardRunes[i]))
		temp = cardBit * secondLineWeight[i]
		result += temp
	}

	secondDigital = 11 - result%11
	if secondDigital > 9 {
		secondDigital = 0
	}
	return fmt.Sprintf("%s%d", cardNo, secondDigital)
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
		log.Println("[ERR] - JSON encode error:", err)
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
