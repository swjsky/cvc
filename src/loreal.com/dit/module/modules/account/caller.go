package account

import (
	"net/http"
)

//GetCaller account from request header
func GetCaller(r *http.Request) *Account {
	caller := &Account{
		UID:   r.Header.Get("uid"),
		Roles: r.Header.Get("roles"),
	}
	if caller.UID == "" {
		caller.UID = "anonymous"
		caller.Roles = "anonymous"
	}
	return caller
}
