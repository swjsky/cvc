package middlewares

import (
	"net/http"
	"strings"

	"loreal.com/dit/endpoint"
)

//FileServerRootRedirect returns a Middlewarte that redirect (remove) prefix from a file server Endpoint
//configured with the given prefix.
func FileServerRootRedirect(prefix string) endpoint.ServerMiddleware {
	return func(c endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, prefix) {
				r.URL.Path = r.URL.Path[len(prefix):]
			}
			c.ServeHTTP(w, r)
		})
	}
}
