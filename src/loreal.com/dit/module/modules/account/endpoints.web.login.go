package account

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

//LoggedInUsers - store a list of logged in users
type LoggedInUsers struct {
	LoginLimit int
	Logins     map[string][]time.Time
	mutex      *sync.Mutex
}

func removeExpires(list *[]time.Time) {
	t := time.Now()
	i := len(*list) - 1
	for i >= 0 && len(*list) > 0 {
		if t.Before((*list)[i]) {
			i--
			continue
		}
		(*list)[i] = (*list)[len(*list)-1]  // Copy last element to index i.
		(*list)[len(*list)-1] = time.Time{} // Erase last element (write zero value).
		(*list) = (*list)[:len((*list))-1]  // Truncate slice.
		i--
	}
}

//Login - login
func (lu *LoggedInUsers) Login(id string, expireIn time.Duration) bool {
	lu.mutex.Lock()
	defer lu.mutex.Unlock()
	expireTimes, found := lu.Logins[id]
	if !found {
		lu.Logins[id] = []time.Time{
			time.Now().Add(expireIn),
		}
		log.Printf("[INFO] - Logged in user: %s, %d/%d", id, len(lu.Logins[id]), lu.LoginLimit)
		return true
	}
	removeExpires(&expireTimes)
	if len(expireTimes) >= lu.LoginLimit {
		log.Printf("[WARN] - Too many logins, uid: %s, %d/%d", id, len(lu.Logins[id]), lu.LoginLimit)
		return false
	}
	lu.Logins[id] = append(expireTimes, time.Now().Add(expireIn))
	log.Printf("[INFO] - Add logged in user: %s, %d/%d", id, len(lu.Logins[id]), lu.LoginLimit)
	return true
}

//Logout - login
func (lu *LoggedInUsers) Logout(id string) {
	lu.mutex.Lock()
	defer lu.mutex.Unlock()
	if expireTimes, found := lu.Logins[id]; found {
		if len(expireTimes) == 1 {
			delete(lu.Logins, id)
		}
		lu.Logins[id] = expireTimes[1:]
	}
	log.Printf("[INFO] - Logged out uid: %s, %d/%d", id, len(lu.Logins[id]), lu.LoginLimit)
}

//SetLimit - SetLimit
func (lu *LoggedInUsers) SetLimit(loginLimit int) {
	lu.mutex.Lock()
	defer lu.mutex.Unlock()
	lu.LoginLimit = loginLimit
}

//webLoginHandler - login
//endpoint: token
//method: GET
func (m *Module) webLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	uid := r.Header.Get("uid")
	accounts := m.GetAccounts(uid)
	if len(accounts) == 0 {
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}
	account := accounts[0]
	if !m.LoggedInUsers.Login(uid, time.Second*m.tokenExpiresIn) {
		http.Error(w, "403", http.StatusForbidden)
		return
	}
	webToken := m.newToken(account)
	cookie := &http.Cookie{
		Name:     DefaultAccountConfig.CookieName,
		Value:    webToken,
		Domain:   DefaultAccountConfig.CookieDomain,
		Path:     DefaultAccountConfig.CookiePath,
		HttpOnly: true,
		Secure:   DefaultAccountConfig.CookieSecure,
		MaxAge:   DefaultAccountConfig.MaxAge,
	}
	debugInfo("account", fmt.Sprintf("Set-Cookie: %s", cookie.String()), 2)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, DefaultAccountConfig.LoginPath, http.StatusMovedPermanently)
}

//webLogoutHandler - logout
//endpoint: token
//method: GET
func (m *Module) webLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	if tokenCookie, err := r.Cookie(m.GetWebTokenCookieName()); err == nil {
		u := m.VerifyToken(tokenCookie.Value)
		if u != nil {
			m.LoggedInUsers.Logout(u.UID)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     DefaultAccountConfig.CookieName,
		Value:    "",
		Domain:   DefaultAccountConfig.CookieDomain,
		Path:     DefaultAccountConfig.CookiePath,
		HttpOnly: true,
		Secure:   DefaultAccountConfig.CookieSecure,
		MaxAge:   -1,
	})
	http.Redirect(w, r, DefaultAccountConfig.LogoutPath, http.StatusMovedPermanently)
}
