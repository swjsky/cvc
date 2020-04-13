//Package endpoint is the fundamental building block for CEH, It enables middlewares for http servers and clients.
package endpoint

import "net/http"

// Endpoint is the fundamental building block of servers and clients.
type Endpoint http.Handler

// Impl is the one implementation for Endpoint.
type Impl http.HandlerFunc

func (e Impl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e(w, r)
}

// HTTPClient is the fundamental building block of servers and clients.
type HTTPClient func(r *http.Request) (resp *http.Response, err error)

//Do send the http request
func (c HTTPClient) Do(r *http.Request) (resp *http.Response, err error) {
	return c(r)
}

// ServerMiddleware is a chainable behavior modifier for endpoints.
type ServerMiddleware func(Endpoint) Endpoint

// ClientMiddleware is a chainable behavior modifier for endpoints.
type ClientMiddleware func(HTTPClient) HTTPClient

//DecorateServer decorates an Endpoint e with all the given Middlewares, in order.
func DecorateServer(e Endpoint, ms ...ServerMiddleware) Endpoint {
	decorated := e
	for i := len(ms) - 1; i >= 0; i-- { // reverse
		decorated = ms[i](decorated)
	}
	return decorated
}

//DecorateClient decorates an ClientEndpoint c with all the given Middlewares, in order.
func DecorateClient(c *http.Client, ms ...ClientMiddleware) HTTPClient {
	decorated := c.Do
	for i := len(ms) - 1; i >= 0; i-- { // reverse
		decorated = ms[i](decorated)
	}
	return decorated
}

//DecorateEndpointClient decorates an ClientEndpoint c with all the given Middlewares, in order.
func DecorateEndpointClient(c *HTTPClient, ms ...ClientMiddleware) HTTPClient {
	decorated := c.Do
	for i := len(ms) - 1; i >= 0; i-- { // reverse
		decorated = ms[i](decorated)
	}
	return decorated
}
