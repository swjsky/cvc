//Package utils - common functions used in CEH project
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

//RandomString generate random string by length
func RandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

//RandomNumber  随机数
func RandomNumber(n int) string {
	rand.Seed(time.Now().UnixNano())
	str := ""
	for i := 0; i < n; i++ {
		str = str + fmt.Sprint(rand.Intn(10))
	}
	return str
}

//LoadOrCreateJSON - load JSON file or create new JSON from defaultData
func LoadOrCreateJSON(fileName string, defaultData interface{}) error {
	var err error
	if err = LoadJSON(fileName, &defaultData); err != nil {
		if os.IsNotExist(err) {
			err = SaveJSON(defaultData, fileName)
			return err
		}
		log.Println("[Load JSON]", fileName, err)
		return err
	}
	return nil
}

//LoadOrCreateJSONNoEscape - load JSON file or create new JSON from defaultData
func LoadOrCreateJSONNoEscape(fileName string, defaultData interface{}) error {
	var err error
	if err = LoadJSON(fileName, &defaultData); err != nil {
		if os.IsNotExist(err) {
			err = SaveJSONNoEscape(defaultData, fileName)
			return err
		}
		log.Println("[Load JSON]", fileName, err)
		return err
	}
	return nil
}

//SaveJSON file
func SaveJSON(data interface{}, fileName string) error {
	var err error
	var b []byte
	var f *os.File
	if f, err = os.Create(fileName); err != nil {
		return err
	}
	defer f.Close()
	if b, err = json.MarshalIndent(data, "", "\t"); err != nil {
		return err
	}
	if _, err = f.Write(b); err != nil {
		return err
	}
	return nil
}

//SaveJSONNoEscape - save to file without escape
func SaveJSONNoEscape(data interface{}, fileName string) error {
	var err error
	var f *os.File
	if f, err = os.Create(fileName); err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(MarshalJSONV2(data, "\t", false)); err != nil {
		return err
	}
	return nil
}

//LoadJSON file
func LoadJSON(fileName string, v interface{}) error {
	var err error
	var b []byte
	var f *os.File
	if f, err = os.Open(fileName); err != nil {
		return err
	}
	defer f.Close()
	if b, err = ioutil.ReadAll(f); err != nil {
		return err
	}
	return json.Unmarshal(b, &v)
}

//Substr 截取子串,这个对中文无效
func Substr(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

//OutputJSON - OutputJSON helper
func OutputJSON(w http.ResponseWriter, errCode int, errMsg string, params ...interface{}) {
	r := map[string]interface{}{"errcode": errCode, "errmsg": errMsg}

	if len(params)%2 == 0 {
		for i := 0; i < len(params); i = i + 2 {
			if key, ok := params[i].(string); ok {
				r[key] = params[i+1]
			}
		}
	}

	w.Header().Set("Content-type", "application/json;charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	b, _ := json.Marshal(r)
	w.Write(b)
}

//GetCardKeys - extract keys from cardnumber
func GetCardKeys(cardno string) (storekey, custkey, cardholderkey int) {
	if len(cardno) != 22 {
		return
	}
	cardholderno := cardno[16:18]
	custno := cardno[8:16]
	storeno := cardno[5:8]

	cardholderkey, _ = strconv.Atoi(cardholderno)
	custkey, _ = strconv.Atoi(custno)
	storekey, _ = strconv.Atoi(storeno)
	return
}

//MakeFolder - make folder if it is not exist.
func MakeFolder(path string) {
	path = NormalizePath(path)
	if err := os.MkdirAll(path, os.ModeDir|0770); err != nil {
		log.Println("[Mkdir]", err)
	}
}

//NormalizePath - Add / at the end of path
func NormalizePath(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

//NormalizePathPtr - Add / at the end of path
func NormalizePathPtr(path *string) {
	if !strings.HasSuffix(*path, "/") {
		*path = *path + "/"
	}
}

//PrintJSON - print indent json string
func PrintJSON(data interface{}) {
	fmt.Println(MarshalJSON(data))
}

//MarshalJSON - marshal indent json string
func MarshalJSON(data interface{}, indent ...bool) string {
	if len(indent) > 0 && indent[0] {
		b, _ := json.MarshalIndent(data, "", "  ")
		return string(b)
	}
	b, _ := json.Marshal(data)
	return string(b)
}

//MarshalJSONV2 - Marshal JSON without html Escape
func MarshalJSONV2(t interface{}, indent string, escape bool) []byte {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	if indent != "" {
		encoder.SetIndent("", indent)
	}
	encoder.SetEscapeHTML(escape)
	if err := encoder.Encode(t); err != nil {
		log.Println("[ERR] - [MarshalJSONNoEscape]:", err)
	}
	return buffer.Bytes()
}
