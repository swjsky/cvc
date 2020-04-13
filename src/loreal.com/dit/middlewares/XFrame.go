package middlewares

import (
	"net/http"

	"loreal.com/dit/endpoint"
)

//XFrame Add X-Frame-options into Header
func XFrame(mode string) endpoint.ServerMiddleware {
	return func(c endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if mode == "" {
				mode = "sameorigin"
			}
			w.Header().Set("X-Frame-Options", mode)
			c.ServeHTTP(w, r)
		})
	}
}
