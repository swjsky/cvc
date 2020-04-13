package oauth

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"


	"loreal.com/dit/endpoint"
	"loreal.com/dit/cmd/coupon-service/base"

	// "github.com/dgrijalva/jwt-go"
)

//OauthCheckToken 验证请求的token 目前比较简单。
func OauthCheckToken() endpoint.ServerMiddleware {
	return func(c endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			// token := r.Header.Get("Authorization")
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if	base.IsBlankString(token) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if	base.IsBlankString(token) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// err := validateToken(token)
			
			// var jwtClaims jwt.StandardClaims 
			ret, err := base.ParesToken(token, base.Pubkey)
			if nil != err {
				if &base.ErrTokenExpired == err {
					base.WriteErrorResponse(w, http.StatusBadRequest, "application/json;charset=utf-8", err)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
			// jwtClaims = (jwt.StandardClaims)ret
			// ccc := ret.(*jwt.Token).Claims
			// r.Header.Add("mapclaims", ret)
			log.Println("解析成功 :", ret)
			// log.Println("将要删除 :", ccc)
			c.ServeHTTP(w, r)
		})
	}
}

// func pares(tokenString string, key interface{}) (interface{}, bool) {
// 	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {

// 		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
// 			return nil, fmt.Errorf("不支持的加密方式: %v", token.Header["alg"])
// 		}
// 		return key, nil
// 	})
// 	if	nil != err {
// 		log.Println("解析token失败:", err)
// 		return nil, false
// 	}

// 	// if claims, ok := token.Claims.(jwt.StandardClaims); ok && token.Valid {
// 	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {	
// 		return claims, true
// 	} else {
// 		log.Println("======pares:", err)
// 		return "", false
// 	}

// }


// func pares(tokenString string, key interface{}) (interface{}, error) {
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

// 		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
// 			return nil, fmt.Errorf("不支持的加密方式: %v", token.Header["alg"])
// 		}
// 		return key, nil
// 	})
// 	if	nil != err {
// 		log.Println("解析token失败:", err)
// 		return nil, false
// 	}

// 	// if claims, ok := token.Claims.(jwt.StandardClaims); ok && token.Valid {
// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {	
// 		return claims, true
// 	} else {
// 		log.Println("======pares:", err)
// 		return "", false
// 	}

// }


// validateToken 向认证服务器check token
// 暂时用arvato的服务器，后面用自己的，目前简单实现，如果没有err就算过了
// 暂时各种hard code,也不check是否active
func validateToken(token string) error {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	err := w.WriteField("hello", "world")
	if err != nil {
		log.Println(err)
		return err
	}

	var u = "http://139.217.239.202/api/interface/oauth/check_token?token=" + token
	req, err := http.NewRequest("POST", u, buf)
	if err != nil {
		log.Println("validateToken req err: ", err)
		return err
	}
	req.Header.Set("Content-Type", "multipart/form-data")
	req.Header.Set("Authorization", "Basic YWRtaW46Y3JpdXNhZG1pbg==")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("validateToken response err: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		rstr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("ioutil ReadAll failed :", err.Error())
		}
		log.Println("resp status:" + fmt.Sprint(resp.StatusCode) + string(rstr))
		return errors.New("resp status:" + fmt.Sprint(resp.StatusCode) + string(rstr))
	}

	return nil
}