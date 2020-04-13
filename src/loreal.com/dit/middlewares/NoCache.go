package middlewares

import (
	"net/http"

	"loreal.com/dit/endpoint"
)

//NoCache returns a Middlewarte that redirect (remove) prefix from a file server Endpoint
//configured with the given prefix.
func NoCache() endpoint.ServerMiddleware {
	return func(c endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
			c.ServeHTTP(w, r)
		})
	}
}
