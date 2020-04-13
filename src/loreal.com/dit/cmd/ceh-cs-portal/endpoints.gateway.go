package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

/* 以下为具体 Endpoint 实现代码 */

//kvstoreHandler - get value from kvstore in runtime
//endpoint: /api/kvstore
//method: GET
func (a *App) gatewayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		outputJSON(w, APIStatus{
			ErrCode:    -1,
			ErrMessage: "not acceptable",
		})
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("[ERR] - [gatewayHandler][ParseForm], err: %v", err)
	}
	path := sanitizePolicy.Sanitize(r.PostFormValue("path"))
	r.PostForm.Del("path")

	payload := map[string]interface{}{}
	//prepare payload
	for k, v := range r.PostForm {
		if v == nil || len(v) == 0 {
			continue
		}
		value := sanitizePolicy.Sanitize(v[0])
		r.PostForm.Set(k, value)
		payload[k] = value
	}
	bodyBuffer := bytes.NewBuffer(nil)
	enc := json.NewEncoder(bodyBuffer)
	if err := enc.Encode(payload); err != nil {
		outputJSON(w, APIStatus{
			ErrCode:    -2,
			ErrMessage: "500",
		})
		return
	}

	req, err := http.NewRequest(http.MethodPost, a.Config.UpstreamURL+path, bytes.NewReader(bodyBuffer.Bytes()))
	if err != nil {
		outputJSON(w, APIStatus{
			ErrCode:    -3,
			ErrMessage: "500",
		})
		return
	}
	account := a.getAPIAccount()
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	accessToken := account.GetToken(false)
retry:
	if accessToken == "" {
		log.Println("[ERR] - [gatewayHandler] Cannot get token")
		outputJSON(w, APIStatus{
			ErrCode:    -4,
			ErrMessage: "500",
		})
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("[ERR] - [gatewayHandler] http.do err:", err)
		outputJSON(w, APIStatus{
			ErrCode:    -5,
			ErrMessage: "502",
		})
		return
	}
	defer resp.Body.Close()
	bodyObject := map[string]interface{}{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bodyObject); err != nil {
		outputJSON(w, APIStatus{
			ErrCode:    -6,
			ErrMessage: "500",
		})
		return
	}
	if errCode, ok := bodyObject["error"]; ok {
		strErr, _ := errCode.(string)
		switch strings.ToLower(strErr) {
		case "invalid_token":
			accessToken = account.GetToken(true)
			goto retry
		case "not found":
			outputJSON(w, APIStatus{
				ErrCode:    -7,
				ErrMessage: "404",
			})
			return
		default:
			msg, _ := bodyObject["error_description"]
			log.Printf("[ERR] - [interface][wg] errcode: %s, msg: %v", strErr, msg)
		}
	}
	//log.Println("[OUTPUT] - ", bodyObject)
	outputJSON(w, bodyObject)
}
