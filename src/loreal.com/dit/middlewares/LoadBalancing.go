package middlewares

import (
	"math/rand"
	"net/http"
	"net/url"
	"sync/atomic"

	"loreal.com/dit/endpoint"
)

//Director modifies an http Request to follow a load balancing strategy.
type Director func(*http.Request)

//LoadBalancing returns a Middlewarte that load balances an Endpoint's requests across
//multiple backends using the given Director.
func LoadBalancing(dir *Director) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			(*dir)(r)
			e.ServeHTTP(w, r)
		})
	}
}

//ClientLoadBalancing returns a Middlewarte that load balances an HTTPClient's requests across
//multiple backends using the given Director.
func ClientLoadBalancing(dir *Director) endpoint.ClientMiddleware {
	return func(c endpoint.HTTPClient) endpoint.HTTPClient {
		return func(r *http.Request) (resp *http.Response, err error) {
			(*dir)(r)
			//log.Println(r)
			return c.Do(r)
		}
	}
}

//RoundRobin returns a Balancer which round-robins across the given backends.
func RoundRobin(robin uint64, backends ...string) Director {
	hosts := make([]string, 0, len(backends))
	for _, urlStr := range backends {
		if url, err := url.Parse(urlStr); err == nil {
			hosts = append(hosts, url.Host)
		}
	}
	//log.Println(hosts)
	return func(r *http.Request) {
		if len(backends) > 0 {
			r.URL.Host = hosts[atomic.AddUint64(&robin, 1)%uint64(len(hosts))]
		}
	}
}

//Random returns a Balancer which randomly picks one of the given backends.
func Random(seed int64, backends ...string) Director {
	hosts := make([]string, 0, len(backends))
	for _, urlStr := range backends {
		if url, err := url.Parse(urlStr); err == nil {
			hosts = append(hosts, url.Host)
		}
	}
	return func(r *http.Request) {
		if len(backends) > 0 {
			r.URL.Host = hosts[rand.Intn(len(hosts))]
		}
	}
}
