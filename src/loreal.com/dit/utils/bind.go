package utils

import (
	"encoding/json"
	"net/http"
)

// Binder binder interface
type Binder interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

var (
	// JSON json binder
	JSON = jsonBinding{}
)

// BindJSON bind json request
func BindJSON(r *http.Request, obj interface{}) error {
	return ShouldBindWith(r, obj, JSON)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine
func ShouldBindWith(r *http.Request, obj interface{}, b Binder) error {
	return b.Bind(r, obj)
}

////////////////////////////////////////////////////////
// JSON binder
type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj interface{}) error {
	return json.NewDecoder(req.Body).Decode(obj)
}
