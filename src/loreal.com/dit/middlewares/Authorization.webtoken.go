package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"loreal.com/dit/endpoint"
)

//WebTokenVerifier - verify and extract info. from webToken
type WebTokenVerifier interface {
	WebTokenValid(webToken string, uid, roles *string, properties, publicProps *[]byte) bool
	GetWebTokenCookieName() string
}

//WebTokenAuth - Auth with web-token by cookie
func WebTokenAuth(v WebTokenVerifier) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			webTokenCookie, err := r.Cookie(v.GetWebTokenCookieName())
			if err != nil {
				http.Error(w, "无权访问", http.StatusForbidden)
				return
			}
			var uid, roles string
			var properties, publicProps []byte
			if !v.WebTokenValid(
				webTokenCookie.Value,
				&uid,
				&roles,
				&properties,
				&publicProps,
			) {
				http.Error(w, "无权访问", http.StatusForbidden)
				return
			}
			r.Header.Set("uid", uid)
			r.Header.Set("roles", roles)
			decodeProperties(r, bytes.NewReader(publicProps))
			decodeProperties(r, bytes.NewReader(properties))

			e.ServeHTTP(w, r)
		})
	}
}

func decodeProperties(r *http.Request, reader io.Reader) error {
	dec := json.NewDecoder(reader)
	var p map[string]string
	if err := dec.Decode(&p); err != nil {
		if err == io.EOF {
			return nil
		}
		log.Printf("[ERR] - [Authorization.webtoken.go][Decode][Properties] %v\n", err)
		return err
	}
	for k, v := range p {
		r.Header.Set(k, v)
	}
	return nil
}
