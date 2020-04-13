package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"loreal.com/dit/endpoint"
)

//VerifyTokenCallback - callback func to verify token
type VerifyTokenCallback func(tokenOk bool, uid, roles string)

//UserReader - read user info from data storage
type UserReader interface {
	ReadUser(username, realm string) (password string, ok bool)
}

//UserVerifier - verify user/password info from data storage
type UserVerifier interface {
	VerifyPassword(username, password, realm string) (ok bool)
}

//UserRoleVerifier - verify user/password and role info from data storage
type UserRoleVerifier interface {
	VerifyPasswordAndRole(username, password, realm, role string, callback VerifyTokenCallback)
}

//TokenRoleVerifier - verify token and role info from data storage
type TokenRoleVerifier interface {
	VerifyTokenAndRole(token, realm, role string, callback VerifyTokenCallback)
}

//RoleVerifier - verify role info by token or user & pass from data storage
type RoleVerifier interface {
	UserRoleVerifier
	TokenRoleVerifier
}

//TokenLocater - find token from data storage
type TokenLocater interface {
	FindToken(token, realm string, callback VerifyTokenCallback)
}

//BasicAuth returns a Middleware that authorizes every Endpoint requests
//with the given token
func BasicAuth(ur UserReader, realm string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); ok {
				if pass, found := ur.ReadUser(u, realm); found {
					if pass == p {
						e.ServeHTTP(w, r)
						return
					}
				}
			}
			w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

//BasicAuthMd5 returns a Middleware that authorizes every Endpoint requests
//with the given token
func BasicAuthMd5(uv UserVerifier, realm string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); ok {
				if passed := uv.VerifyPassword(u, p, realm); passed {
					e.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

//BasicAuthWithRole returns a Middleware that authorizes every Endpoint requests
//with the given token
func BasicAuthWithRole(rv UserRoleVerifier, realm, role string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); ok {
				rv.VerifyPasswordAndRole(u, p, realm, role, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
				return
			}
			w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

//BasicAuthWithRoleSkipable returns a skipable Middleware that authorizes every Endpoint requests
//with the given token, or uid and password
func BasicAuthWithRoleSkipable(rv RoleVerifier, realm, role string, skipAuth bool) endpoint.ServerMiddleware {
	if skipAuth {
		return skipAuthFunc()
	}
	return BasicAuthWithRole(rv, realm, role)
}

//BasicAuthOrTokenAuthWithRole returns a Middleware that authorizes every Endpoint requests
//with the given token, or uid and password
func BasicAuthOrTokenAuthWithRole(rv RoleVerifier, realm, role string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); ok {
				rv.VerifyPasswordAndRole(u, p, realm, role, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
				return
			} else if authData := r.Header.Get("Authorization"); strings.HasPrefix(authData, "Bearer ") {
				token := authData[7:]
				rv.VerifyTokenAndRole(token, realm, role, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
				return
			}
			w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

//TokenAuthWithRole returns a Middleware that authorizes every Endpoint requests
//with the given token
func TokenAuthWithRole(rv RoleVerifier, realm, role string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if authData := r.Header.Get("Authorization"); strings.HasPrefix(authData, "Bearer ") {
				token := authData[7:]
				rv.VerifyTokenAndRole(token, realm, role, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
				return
			} else if token := r.URL.Query().Get("token"); token != "" {
				rv.VerifyTokenAndRole(token, realm, role, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized,Token invalid", http.StatusUnauthorized)
				})
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func skipAuthFunc() endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("uid", "anonymous")
			r.Header.Add("roles", "anonymous")
			e.ServeHTTP(w, r)
		})
	}
}

//BasicAuthOrTokenAuthWithRoleSkipable returns a Middleware that authorizes every Endpoint requests
//with the given token, or uid and password
func BasicAuthOrTokenAuthWithRoleSkipable(rv RoleVerifier, realm, role string, skipAuth bool) endpoint.ServerMiddleware {
	if skipAuth {
		return skipAuthFunc()
	}
	return BasicAuthOrTokenAuthWithRole(rv, realm, role)
}

//BearerAuth defines an authorization bearer token header in the incomming request
func BearerAuth(tl TokenLocater, realm string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if authData := r.Header.Get("Authorization"); strings.HasPrefix(authData, "Bearer ") {
				token := authData[7:]
				tl.FindToken(token, realm, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

//BearerAuthWithRole defines an authorization bearer token header in the incomming request
func BearerAuthWithRole(trv TokenRoleVerifier, realm, role string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if authData := r.Header.Get("Authorization"); strings.HasPrefix(authData, "Bearer ") {
				token := authData[7:]
				trv.VerifyTokenAndRole(token, realm, role, func(tokenOk bool, uid, roles string) {
					if tokenOk {
						r.Header.Add("uid", uid)
						r.Header.Add("roles", roles)
						e.ServeHTTP(w, r)
						return
					}
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				})
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

//TokenAuth defines an authorization by token
func TokenAuth(tl TokenLocater, realm string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			tl.FindToken(token, realm, func(tokenOk bool, uid, roles string) {
				if tokenOk {
					r.Header.Add("uid", uid)
					r.Header.Add("roles", roles)
					e.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			})
		})
	}
}

//Header returns a Middlewarte that adds given HTTP header to every requests
//done by a Endpoint
func Header(name, value string) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add(name, value)
			e.ServeHTTP(w, r)
		})
	}
}

//ClientBasicAuth returns a Middleware that authorizes every HTTPClient requests
//with the given token
func ClientBasicAuth(username, password string) endpoint.ClientMiddleware {
	return func(c endpoint.HTTPClient) endpoint.HTTPClient {
		return func(r *http.Request) (resp *http.Response, err error) {
			r.SetBasicAuth(username, password)
			return c.Do(r)
		}
	}
}

//ClientBearerAuth defines an authorization bearer token header in the outgoing request
func ClientBearerAuth(token string) endpoint.ClientMiddleware {
	return ClientHeader("Authorization", "Bearer "+token)
}

//ClientHeader returns a Middlewarte that adds given HTTP header to every requests
//done by a Endpoint
func ClientHeader(name, value string) endpoint.ClientMiddleware {
	return func(c endpoint.HTTPClient) endpoint.HTTPClient {
		return func(r *http.Request) (resp *http.Response, err error) {
			r.Header.Add(name, value)
			return c.Do(r)
		}
	}
}
