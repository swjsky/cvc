package middlewares

import (
	"net/http"

	"loreal.com/dit/endpoint"
)

//XSS Add X-XSS-protection & X-Content-Type-Options:nosniff Header
func XSS(mode string) endpoint.ServerMiddleware {
	return func(c endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if mode == "" {
				mode = "1; mode=block"
			}
			w.Header().Set("X-XSS-Protection", mode)
			// Add X-Content-Type-Options header
			w.Header().Add("X-Content-Type-Options", "nosniff")
			// Add Content-Security-Policy header
			w.Header().Add("Content-Security-Policy", "script-src 'self'")
			c.ServeHTTP(w, r)
		})
	}
}
