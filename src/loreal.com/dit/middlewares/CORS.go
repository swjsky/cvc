package middlewares

import (
	"net/http"

	"loreal.com/dit/endpoint"
)

//CORS returns a Middlewarte that handle CORS
func CORS(acceptedOrigin, AcceptedMethods, AcceptedHeaders, MaxAge string) endpoint.ServerMiddleware {
	return func(c endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", acceptedOrigin)
			if AcceptedMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", AcceptedMethods)
			}
			if AcceptedHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", AcceptedHeaders)
			}
			if MaxAge != "" {
				w.Header().Set("Access-Control-Max-Age", MaxAge)
			}
			if r.Method != "OPTIONS" {
				c.ServeHTTP(w, r)
			}
		})
	}
}
