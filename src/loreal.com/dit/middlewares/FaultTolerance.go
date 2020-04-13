package middlewares

import (
	"net/http"
	"time"

	"loreal.com/dit/endpoint"
)

//FaultTolerance returns a Middlewarte that extents a Endpoint with fault tolerance
//configured with the given attempts and backoff duration.
func FaultTolerance(attempts int, backoff time.Duration, inspectCallback func(resp *http.Response) bool) endpoint.ClientMiddleware {
	return func(c endpoint.HTTPClient) endpoint.HTTPClient {
		return func(r *http.Request) (resp *http.Response, err error) {
			for i := 0; i < attempts; i++ {
				if resp, err = c.Do(r); err == nil {
					break
				}
				if inspectCallback != nil && inspectCallback(resp) {
					break
				}
				time.Sleep(backoff * time.Duration(i))
			}
			return
		}
	}
}
