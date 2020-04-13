package account

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/microcosm-cc/bluemonday"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
)

var sanitizePolicy *bluemonday.Policy

func init() {
	sanitizePolicy = bluemonday.UGCPolicy()
}

//RegisterEndpoints for SMS verify service
func (m *Module) registerEndpoints() {

	m.MountingPoints[""] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			uid := sanitizePolicy.Sanitize(r.URL.Query().Get("uid"))
			caller := GetCaller(r)

			if !caller.IsSelfOrAdmin(uid) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			// var result interface{}

			accounts := m.GetAccounts(uid)
			// result = accounts
			// if len(accounts) == 1 {
			// 	result = accounts[0]
			// }

			w.Header().Set("Content-type", "application/json;charset=utf-8")
			if os.Getenv("EV_DEBUG") != "" {
				b, _ := json.MarshalIndent(accounts, "", "  ")
				w.Write(b)
			} else {
				b, _ := json.Marshal(accounts)
				w.Write(b)
			}
		}),
		middlewares.ServerInstrumentation("profile", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRole(m, "", "admin,user"),
	)

	m.MountingPoints["login"] = endpoint.DecorateServer(
		endpoint.Impl(m.webLoginHandler),
		middlewares.CORS(DefaultAccountConfig.CookieDomain, "GET,POST", "Content-Type,Accept", ""),
		middlewares.ServerInstrumentation("login", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRole(m, "", "admin,user"),
	)

	m.MountingPoints["logout"] = endpoint.DecorateServer(
		endpoint.Impl(m.webLogoutHandler),
		middlewares.ServerInstrumentation("logout", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	)

	m.MountingPoints["public-profile"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			uid := sanitizePolicy.Sanitize(r.URL.Query().Get("uid"))

			account := m.getPublicProps(uid)
			if account == nil {
				http.Error(w, "404", http.StatusNotFound)
				return
			}

			result := make(map[string]interface{})
			result["uid"] = uid
			result["publicprops"] = account.PublicProps

			w.Header().Set("Content-type", "application/json;charset=utf-8")
			if os.Getenv("EV_DEBUG") != "" {
				b, _ := json.MarshalIndent(result, "", "  ")
				w.Write(b)
			} else {
				b, _ := json.Marshal(result)
				w.Write(b)
			}
		}),
		middlewares.ServerInstrumentation("public-profile", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRole(m, "", "admin,user,anonymous"),
	)

	m.MountingPoints["token"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			password := r.PostForm.Get("pass")
			withProperties := r.PostForm.Get("profile")
			account, err := m.Authenticate(uid, password)
			switch {
			case err == ErrAccountLocked:
				http.Error(w, fmt.Sprint("Unauthorized, Account locked!"), http.StatusUnauthorized)
				break
			case err == nil && account != nil:
				token := m.newToken(account)
				expiresIn := -1
				if uid != "anonymous" {
					expiresIn = int(m.tokenExpiresIn.Seconds())
				}
				w.Header().Set("Content-type", "application/json;charset=utf-8")
				result := TokenPackage{
					Token:     token,
					ExpiresIn: expiresIn,
				}
				if withProperties != "" {
					result.Roles = account.Roles
					account.ParseProperties()
					account.ParsePublicProps()
					result.Properties = account.Properties
					result.PublicProps = account.PublicProps
				}

				if b, err := json.MarshalIndent(result, "", " "); err == nil {
					w.Write(b)
				} else {
					log.Println("[ERR][EP][token]", err)
					http.Error(w, "500", http.StatusInternalServerError)
				}
				break
			default:
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				break
			}
		}),
		middlewares.ServerInstrumentation("account-token", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.NoCache(),
	)

	m.MountingPoints["init"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			q := r.URL.Query()
			password := q.Get("pass")
			if m.init(password) {
				http.Error(w, "OK", http.StatusOK)
			} else {
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		}),
		middlewares.ServerInstrumentation("account-init", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	)

	m.MountingPoints["add"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			password := r.PostForm.Get("pass")
			roles := r.PostForm.Get("roles")
			strProperties := r.PostForm.Get("properties")
			strPublicProps := r.PostForm.Get("publicprops")
			if strProperties == "" {
				strProperties = "{}"
			}
			if strPublicProps == "" {
				strPublicProps = "{}"
			}
			properties := make(map[string]interface{})
			publicprops := make(map[string]interface{})

			if err := json.Unmarshal([]byte(strProperties), &properties); err != nil {
				log.Printf("Invalid properties: %v", err)
				http.Error(w, "Invalid properties", http.StatusInternalServerError)
				return
			}
			if err := json.Unmarshal([]byte(strPublicProps), &publicprops); err != nil {
				log.Printf("Invalid properties: %v", err)
				http.Error(w, "Invalid properties", http.StatusInternalServerError)
				return
			}

			userAccount := &Account{
				UID:   uid,
				Roles: roles,
			}
			if err := userAccount.setPassword(password); err != nil {
				log.Println("[DEBUG] - p0")
				http.Error(w, "password not acceptable", http.StatusNotAcceptable)
				return
			}
			userAccount.SetProperties(properties)
			userAccount.SetPublicProps(publicprops)
			if err := m.insertAccount(userAccount); err == nil {
				w.Write([]byte("OK"))
				return
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}),
		middlewares.ServerInstrumentation("create-account", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BearerAuthWithRole(m, "", "admin"),
	)

	m.MountingPoints["remove"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			if uid == "admin" {
				http.Error(w, "Can not delete admin account", http.StatusForbidden)
				return
			}
			if err := m.removeAccount(uid); err == nil {
				http.Error(w, "OK", http.StatusOK)
			} else {
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		}),
		middlewares.ServerInstrumentation("remove-account", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BearerAuthWithRole(m, "", "admin"),
	)

	m.MountingPoints["register"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			password := r.PostForm.Get("pass")
			strProperties := r.PostForm.Get("properties")
			strPublicProps := r.PostForm.Get("publicprops")
			if strProperties == "" {
				strProperties = "{}"
			}
			if strPublicProps == "" {
				strPublicProps = "{}"
			}
			properties := make(map[string]interface{})
			publicprops := make(map[string]interface{})

			if err := json.Unmarshal([]byte(strProperties), &properties); err != nil {
				log.Printf("Invalid properties: %v", err)
				http.Error(w, "Invalid properties", http.StatusInternalServerError)
				return
			}
			if err := json.Unmarshal([]byte(strPublicProps), &publicprops); err != nil {
				log.Printf("Invalid properties: %v", err)
				http.Error(w, "Invalid properties", http.StatusInternalServerError)
				return
			}
			roles := "user"
			userAccount := &Account{
				UID:   uid,
				Roles: roles,
			}
			if err := userAccount.setPassword(password); err != nil {
				http.Error(w, "password not acceptable", http.StatusNotAcceptable)
				return
			}
			userAccount.SetProperties(properties)
			userAccount.SetPublicProps(publicprops)

			if err := m.insertAccount(userAccount); err == nil {
				http.Error(w, "OK", http.StatusOK)
			} else {
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		}),
		middlewares.ServerInstrumentation("register-account", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BearerAuthWithRole(m, "", "anonymous"),
	)

	m.MountingPoints["update"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			targetUID := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			roles := r.PostForm.Get("roles")
			strProperties := r.PostForm.Get("properties")
			strPublicProps := r.PostForm.Get("publicprops")

			caller := GetCaller(r)
			target := m.GetAccount(targetUID)
			if target == nil {
				http.Error(w, "404", http.StatusNotFound)
				return
			}

			if target.IsAnotherAdmin(caller.UID) {
				log.Printf("[WARN] - [EP][update] Cannot update another admin account: caller %s, target: %s\n", caller.UID, target.UID)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if !caller.IsSelfOrAdmin(targetUID) {
				log.Printf("[WARN] - [EP][update] Forbidden: caller %s, target: %s\n", caller.UID, target.UID)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if strProperties == "" {
				strProperties = "{}"
			}
			if strPublicProps == "" {
				strPublicProps = "{}"
			}
			properties := make(map[string]interface{})
			publicprops := make(map[string]interface{})

			if err := json.Unmarshal([]byte(strProperties), &properties); err != nil {
				log.Printf("Invalid properties: %v", err)
				http.Error(w, "Invalid properties", http.StatusInternalServerError)
				return
			}
			if err := json.Unmarshal([]byte(strPublicProps), &publicprops); err != nil {
				log.Printf("Invalid properties: %v", err)
				http.Error(w, "Invalid properties", http.StatusInternalServerError)
				return
			}

			target.SetProperties(properties)
			target.SetPublicProps(publicprops)

			if caller.IsAdmin() {
				target.Roles = roles
				if err := m.updateAccount(target); err == nil {
					http.Error(w, "OK", http.StatusOK)
				} else {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			} else {
				if err := m.updateProperties(target); err == nil {
					http.Error(w, "OK", http.StatusOK)
				} else {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			}

		}),
		middlewares.ServerInstrumentation("update-account", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BearerAuthWithRole(m, "", "admin,user"),
	)

	m.MountingPoints["cpw"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			operator := sanitizePolicy.Sanitize(r.PostForm.Get("admin"))
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			oldPassword := r.PostForm.Get("old-pass")
			newPassword1 := r.PostForm.Get("new-pass1")
			newPassword2 := r.PostForm.Get("new-pass2")
			if operator == "" {
				operator = uid
			}
			account, err := m.Authenticate(operator, oldPassword)
			switch {
			case err == ErrAccountLocked:
				http.Error(w, fmt.Sprint("Unauthorized, Account locked!"), http.StatusUnauthorized)
				return
			case err == nil && account != nil:
				caller := GetCaller(r)

				if !caller.IsSelfOrAdmin(uid) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}

				var targetUser = caller

				if caller.IsAdmin() && uid != "" {
					targetUser = &Account{
						UID: uid,
					}
				}
				switch targetUser.SetPassword(newPassword1, newPassword2) {
				case ErrPasswordNotMatch:
					http.Error(w, "password not match", http.StatusNotAcceptable)
					return
				case ErrPasswordRule:
					http.Error(w, "password not acceptable", http.StatusNotAcceptable)
					return
				}
				if err := m.updatePassword(targetUser); err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				w.Write([]byte("OK"))
				return
			default:
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}),
		middlewares.ServerInstrumentation("change-password", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BearerAuthWithRole(m, "", "admin,user"),
	)

	m.MountingPoints["resetpw"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			target := m.GetAccount(uid)
			if target == nil {
				http.Error(w, "404", http.StatusNotFound)
				return
			}
			emailAddr := AsString(GetPropertie(target.PublicProps, "email"))
			if emailAddr == "" {
				http.Error(w, "not acceptable", http.StatusNotAcceptable)
				return
			}
			if err := m.resetPassword(target, emailAddr); err != nil {
				log.Println("[ERR] - [SendMail] err:", err)
				http.Error(w, "500", http.StatusInternalServerError)
				return
			}
			http.Error(w, "OK", http.StatusOK)
		}),
		middlewares.ServerInstrumentation("reset-pw", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRole(m, "", "admin"),
	)

	m.MountingPoints["unlock"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			if err := r.ParseForm(); err != nil {
				log.Printf("Incompatible request: %v", err)
				http.Error(w, "Incompatible request", http.StatusInternalServerError)
				return
			}
			uid := sanitizePolicy.Sanitize(r.PostForm.Get("uid"))
			if err := m.unlockAccount(&Account{UID: uid}); err != nil {
				log.Println("[ERR] - [EP][unlock] err:", err)
				http.Error(w, "500", http.StatusInternalServerError)
				return
			}
			http.Error(w, "OK", http.StatusOK)
		}),
		middlewares.ServerInstrumentation("unlock", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRole(m, "", "admin"),
	)

	// m.MountingPoints["shutdown"] = endpoint.DecorateServer(
	// 	endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
	// 		if r.Method != "GET" {
	// 			http.Error(w, "Not Supported", http.StatusNotAcceptable)
	// 			return
	// 		}
	// 		q := r.URL.Query()
	// 		shutdownKey := q.Get("key")
	// 		if !m.verifyShutdownKey(shutdownKey) {
	// 			http.Error(w, m.setShutdownKey(utils.RandomString(16)), http.StatusOK)
	// 			return
	// 		}
	// 		m.Shutdown()
	// 		http.Error(w, "OK", http.StatusOK)
	// 	}),
	// 	middlewares.BasicAuthMd5(m, "rootshutdown"),
	// )

	// m.MountingPoints["restart"] = endpoint.DecorateServer(
	// 	endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
	// 		if r.Method != "GET" {
	// 			http.Error(w, "Not Supported", http.StatusNotAcceptable)
	// 			return
	// 		}
	// 		m.Restart()
	// 		http.Error(w, "OK", http.StatusOK)
	// 	}),
	// 	middlewares.BasicAuthMd5(m, "rootrestart"),
	// )

}
