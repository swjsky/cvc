package middlewares

import (
	"log"
	"net/http"

	"loreal.com/dit/endpoint"
)

//Logging returns a Middlewarte that logs a Endpoint's requests
func Logging(l *log.Logger) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			l.Printf("%s: %s %s", r.UserAgent(), r.Method, r.URL)
			e.ServeHTTP(w, r)
		})
	}
}

//ClientLogging returns a Middlewarte that logs a Endpoint's requests
func ClientLogging(l *log.Logger) endpoint.ClientMiddleware {
	return func(c endpoint.HTTPClient) endpoint.HTTPClient {
		return func(r *http.Request) (resp *http.Response, err error) {
			l.Printf("%s: %s %s", r.UserAgent(), r.Method, r.URL)
			return c.Do(r)
		}
	}
}
