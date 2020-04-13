package base

import ( 
	// "bytes"
	"encoding/json"
	// "errors"
	"fmt"
	"log"
	"math/rand"
	// "mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/dgrijalva/jwt-go"
)

var r *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))

// IsBlankString 判断是否为空的ID，ID不能都是空白.
func IsBlankString(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// IsEmptyString 判断是否为空的字符串.
func IsEmptyString(str string) bool {
	return len(str) == 0
}

// IsValidUUID
func IsValidUUID(u string) bool {
    _, err := uuid.Parse(u)
    return err == nil
 }
// SetResponseHeader 一个快捷设置status code 和content type的方法
func SetResponseHeader(w http.ResponseWriter, statusCode int, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
}

// WriteErrorResponse 一个快捷设置包含错误body的response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, contentType string, a interface{}) {
	SetResponseHeader(w, statusCode, contentType)
	switch vv := a.(type) {
	case error: {
		fmt.Fprintf(w, vv.Error())
	}
	case map[string][]error: {
		jsonBytes, err := json.Marshal(vv)
		if	nil != err {
			log.Println(err)  
			fmt.Fprintf(w, err.Error())
		} 
		var str = string(jsonBytes)
		fmt.Fprintf(w, str)
	}
	case []error: {
		jsonBytes, err := json.Marshal(vv)
		if	nil != err {
			log.Println(err)  
			fmt.Fprintf(w, err.Error())
		} 
		var str = string(jsonBytes)
		fmt.Fprintf(w, str)
	}
	}
}

// trimURIPrefix 将一个uri拆分为若干node，根据ndoe取得一些动态参数。
func TrimURIPrefix(uri string, stopTag string) []string {
	params := strings.Split(strings.TrimPrefix(strings.TrimSuffix(uri, "/"), "/"), "/")
	last := len(params) - 1
	for i := last; i >= 0; i-- {
		if params[i] == stopTag {
			return params[i+1:]
		}
	}
	return params
}

//outputJSON - output json for http response
func outputJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		log.Println("[ERR] - [outputJSON] JSON encode error:", err)
		http.Error(w, "500", http.StatusInternalServerError)
		return
	}
}

// GetMapClaim 获取JWT 里面的map claims
func GetMapClaim(token string) (interface{}) {
	if	IsBlankString(token) {
		return nil
	}
	ret, b := ParesToken(token, Pubkey)
	if	nil != b {
		return nil
	}
	return ret
}

// ParesToken 检查toekn
func ParesToken(tokenString string, key interface{}) (interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Requester{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("不支持的加密方式: %v", token.Header["alg"])
		}
		return key, nil
	})
	if	nil != err {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				return nil, &ErrTokenExpired
			} 
		}

		log.Println("解析token失败:", err)
		return nil, &ErrTokenValidateFailed
	}

	// if claims, ok := token.Claims.(jwt.StandardClaims); ok && token.Valid {
	if claims, ok := token.Claims.(*Requester); ok && token.Valid {	
		return claims, nil
	} else {
		fmt.Println("======pares:", err)
		return "", &ErrTokenValidateFailed
	}
}
	
// RandString 生成随机字符串
func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}