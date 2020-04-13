package restful

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// var seededRand *rand.Rand
var sanitizePolicy *bluemonday.Policy

var jsonInvalidMethod = map[string]interface{}{
	"errcode": -1,
	"message": "Invalid method",
}
var jsonInvalidData = map[string]interface{}{
	"errcode": -2,
	"message": "Invalid Data",
}

var jsonInvalidID = map[string]interface{}{
	"errcode": -3,
	"message": "Invalid ID",
}

var jsonOPError = map[string]interface{}{
	"errcode": -4,
	"message": "Operation Error",
}

var jsonOK = map[string]interface{}{
	"message": "OK",
}

func init() {
	// 	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	sanitizePolicy = bluemonday.UGCPolicy()
}

//Handler - http handler for a restful endpoint
type Handler struct {
	Name   string /*Endpoint name*/
	Model  interface{}
	Filter func(*http.Request, *map[string]interface{}) bool
}

//NewHandler - create a new instance of RestfulHandler
func NewHandler(name string, model interface{}) *Handler {
	handler := &Handler{
		Name:  name,
		Model: model,
	}
	return handler
}

//SetFilter - set filter
func (h *Handler) SetFilter(filter func(*http.Request, *map[string]interface{}) bool) *Handler {
	if filter != nil {
		h.Filter = filter
	}
	return h
}

//sanitize parameters
func sanitize(params *url.Values) {
	for key := range *params {
		(*params).Set(key, sanitizePolicy.Sanitize((*params).Get(key)))
	}
}

func trimURIPrefix(uri string, stopTag string) []string {
	params := strings.Split(strings.TrimPrefix(strings.TrimSuffix(uri, "/"), "/"), "/")
	last := len(params) - 1
	for i := last; i >= 0; i-- {
		if params[i] == stopTag {
			return params[i+1:]
		}
	}
	return params
}

func parseID(s string) int64 {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}
	return id
}

func (h *Handler) httpGet(w http.ResponseWriter, r *http.Request, id int64) {
	m, ok := h.Model.(Querier)
	if !ok {
		outputGzipJSON(w, jsonInvalidMethod)
		return
	}
	if id != -1 {
		outputGzipJSON(w, map[string]interface{}{
			"message": "ok",
			"method":  "one",
			"payload": m.FindByID(id),
		})
		return
	}
	query := r.URL.Query()
	sanitize(&query)
	total, records := m.Find(query)
	if h.Filter == nil {
		outputGzipJSON(w, map[string]interface{}{
			"message": "ok",
			"method":  "query",
			"total":   total,
			"payload": records,
		})
		return
	}
	finalRecords := make([]*map[string]interface{}, 0, len(records))
	for _, record := range records {
		if !h.Filter(r, record) {
			finalRecords = append(finalRecords, record)
		}
	}
	outputGzipJSON(w, map[string]interface{}{
		"message": "ok",
		"method":  "query",
		"total":   len(finalRecords),
		"payload": finalRecords,
	})
}

func (h *Handler) httpPost(w http.ResponseWriter, r *http.Request, id int64) {
	m, ok := h.Model.(Inserter)
	if !ok {
		outputGzipJSON(w, jsonInvalidMethod)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println("[ERR] - [RestfulHandler][POST][ParseForm] err:", err)
		outputGzipJSON(w, jsonInvalidData)
		return
	}
	sanitize(&r.PostForm)
	newID, err := m.Insert(r.PostForm)
	if err != nil {
		log.Println("[ERR] - [RestfulHandler][POST] err:", err)
		outputGzipJSON(w, jsonOPError)
		return
	}
	outputGzipJSON(w, map[string]interface{}{
		"message": "ok",
		"method":  "insert",
		"id":      newID,
	})
}

func (h *Handler) httpPut(w http.ResponseWriter, r *http.Request, id int64) {
	if err := r.ParseForm(); err != nil {
		log.Println("[ERR] - [RestfulHandler][PUT][ParseForm] err:", err)
		outputGzipJSON(w, jsonInvalidData)
		return
	}
	sanitize(&r.PostForm)

	switch id {
	case -1 /*update by query condition*/ :
		// m, ok := h.Model.(Updater)
		// if !ok {
		// 	outputGzipJSON(w, jsonInvalidMethod)
		// 	return
		// }
		// query := r.URL.Query()
		// sanitize(&query)
		// rowsAffected, err := m.Update(r.PostForm, query)
		// if err != nil {
		// 	log.Println("[ERR] - [RestfulHandler][PUT-Update] err:", err)
		// 	outputGzipJSON(w, jsonOPError)
		// 	return
		// }
		// outputGzipJSON(w, map[string]interface{}{
		// 	"message": "ok",
		// 	"method":  "update",
		// 	"count":   rowsAffected,
		// })
		outputGzipJSON(w, jsonInvalidID)
		return
	default /*update by ID*/ :
		m, ok := h.Model.(Setter)
		if !ok {
			outputGzipJSON(w, jsonInvalidMethod)
			return
		}
		if err := m.Set(id, r.PostForm); err != nil {
			log.Println("[ERR] - [RestfulHandler][PUT-Set] err:", err)
			outputGzipJSON(w, jsonOPError)
			return
		}
		outputGzipJSON(w, map[string]interface{}{
			"message": "ok",
			"method":  "set",
		})
		return
	}
}

func (h *Handler) httpDelete(w http.ResponseWriter, r *http.Request, id int64) {
	m, ok := h.Model.(Deleter)
	if !ok {
		outputGzipJSON(w, jsonInvalidMethod)
		return
	}
	switch id {
	case -1:
		outputGzipJSON(w, jsonInvalidID)
		return
	}
	rowsAffected, err := m.Delete(id)
	if err != nil {
		log.Println("[ERR] - [RestfulHandler][DELETE] err:", err)
		outputGzipJSON(w, jsonOPError)
		return
	}
	outputGzipJSON(w, map[string]interface{}{
		"message": "ok",
		"method":  "delete",
		"count":   rowsAffected,
	})
}

//ServeHTTP - implementation of http.handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Model == nil {
		outputGzipJSON(w, jsonInvalidMethod)
		return
	}
	if DEBUG {
		log.Println("[DEBUG] - [r.RequestURI]:", r.RequestURI)
	}
	params := trimURIPrefix(r.RequestURI, h.Name)
	var id int64 = -1
	if len(params) > 0 {
		id = parseID(sanitizePolicy.Sanitize(params[0]))
	}
	switch r.Method {
	case "GET":
		h.httpGet(w, r, id)
		return
	case "POST":
		h.httpPost(w, r, id)
		return
	case "PUT":
		h.httpPut(w, r, id)
		return
	case "DELETE":
		h.httpDelete(w, r, id)
		return
	default:
		outputGzipJSON(w, jsonInvalidMethod)
	}
}
