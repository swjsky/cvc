package endpoint

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
)

//SendJSON data into body
func SendJSON(w http.ResponseWriter, JSONData interface{}, indent bool) (err error) {
	w.Header().Set("Content-Type", "application/json")
	var b []byte
	if indent {
		b, err = json.MarshalIndent(JSONData, "", " ")
	} else {
		b, err = json.Marshal(JSONData)
	}
	if err != nil {
		return
	}
	w.Write(b)
	return
}

//SendXML data into body
func SendXML(w http.ResponseWriter, XMLData interface{}, indent bool) (err error) {
	w.Header().Set("Content-Type", "text/xml")
	var b []byte
	if indent {
		b, err = xml.MarshalIndent(XMLData, "", " ")
	} else {
		b, err = xml.Marshal(XMLData)
	}
	if err != nil {
		return
	}
	w.Write(b)
	return
}

//ParseJSONRequest Parse Json object from body
func ParseJSONRequest(r *http.Request) (JSONData interface{}, err error) {
	if r.Body == nil {
		err = errors.New("missing JSON body")
		return
	}
	ct := r.Header.Get("Content-Type")
	ct, _, err = mime.ParseMediaType(ct)
	if ct != "application/json" {
		err = errors.New("[ParseJSON] Content-Type mismatch")
		return
	}
	JSONData, err = ParseJSONBody(r.Body)
	return
}

//ParseJSONBody Parse Json object from body
func ParseJSONBody(body io.ReadCloser) (JSONData interface{}, err error) {
	b, e := ioutil.ReadAll(body)
	defer body.Close()
	if e != nil {
		if err == nil {
			err = e
		}
		return
	}
	//先去掉golang包不能处理的字符,只处理32控制字符
	for i, ch := range b {
		switch {
		case ch == '\r':
		case ch == '\n':
		case ch == '\t':
		case ch < ' ':
			b[i] = ' '
		}
	}

	json.Unmarshal(b, &JSONData)
	// m := JSONData.(map[string]interface{})
	// for k, v := range m {
	// 	switch vv := v.(type) {
	// 	case string:
	// 		fmt.Println(k, "is string", vv)
	// 	case int:
	// 		fmt.Println(k, "is int", vv)
	// 	case float64:
	// 		fmt.Println(k, "is float64", vv)
	// 	case []interface{}:
	// 		fmt.Println(k, "is an array:")
	// 		for i, u := range vv {
	// 			fmt.Println(i, u)
	// 		}
	// 	default:
	// 		fmt.Println(k, "is of a type I don't know how to handle")
	// 	}
	// }
	return
}

//ParseXML Parse Json object from body
func ParseXML(r *http.Request) (XMLData interface{}, err error) {
	if r.Body == nil {
		err = errors.New("missing XML body")
		return
	}
	ct := r.Header.Get("Content-Type")
	ct, _, err = mime.ParseMediaType(ct)
	if ct != "text/xml" {
		err = errors.New("[ParseXML] Content-Type mismatch")
		return
	}
	b, e := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if e != nil {
		if err == nil {
			err = e
		}
		return
	}
	xml.Unmarshal(b, &XMLData)
	return
}

//Retry - call function x number of times in case of error
func Retry(times int, errMessage string, funcCanFail func() error) {
	var (
		retryCnt int
		err      error
	)
	for i := 0; i < times; i++ {
		err = funcCanFail()
		if err == nil {
			return
		}
		log.Println(err)
		log.Println(errMessage)
		retryCnt++
		log.Printf("Retry %d ...", retryCnt)
	}
	log.Println(err)
	log.Println("Too many errors, stop trying.")
}

func ParseJSONRequestV2(r *http.Request, v interface{}) (err error) {
	if r.Body == nil {
		err = errors.New("missing JSON body")
		return
	}

	ct := r.Header.Get("Content-Type")
	ct, _, err = mime.ParseMediaType(ct)
	if ct != "application/json" {
		err = errors.New("[ParseJSON] Content-Type mismatch")
		return
	}
	err = ParseJSONBodyV2(r.Body, v)
	return
}

//ParseJSONBody Parse Json object from body
func ParseJSONBodyV2(body io.ReadCloser, v interface{}) (err error) {
	b, e := ioutil.ReadAll(body)
	defer body.Close()
	if e != nil {
		if err == nil {
			err = e
		}
		return
	}
	//先去掉golang包不能处理的字符,只处理32控制字符
	for i, ch := range b {
		switch {
		case ch == '\r':
		case ch == '\n':
		case ch == '\t':
		case ch < ' ':
			b[i] = ' '
		}
	}
	json.Unmarshal(b, v)
	return
}

func ParseXMLV2(r *http.Request, v interface{}) (err error) {
	if r.Body == nil {
		err = errors.New("missing XML body")
		return
	}

	b, e := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if e != nil {
		if err == nil {
			err = errors.New(fmt.Sprintf("io read error: %s", e.Error()))
		}
		return
	}
	xml.Unmarshal(b, v)
	return
}
