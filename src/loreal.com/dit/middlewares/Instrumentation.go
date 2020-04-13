package middlewares

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"loreal.com/dit/endpoint"
)

// ServerInstrumentation returns a Middlewarte that instruments an Endpoint with given metrics
func ServerInstrumentation(serviceName string, requests *prometheus.CounterVec, latency *prometheus.HistogramVec, durationSummary *prometheus.SummaryVec) endpoint.ServerMiddleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			defer func(start time.Time) {
				d := float64(time.Since(start).Nanoseconds())
				latency.WithLabelValues(serviceName).Observe(d)
				durationSummary.WithLabelValues(serviceName).Observe(d)
				requests.WithLabelValues(serviceName).Add(1)
			}(time.Now())
			e.ServeHTTP(w, r)
		})
	}
}

//ClientInstrumentation returns a Middlewarte that instruments a HTTPClient with given metrics
func ClientInstrumentation(serviceName string, requests *prometheus.CounterVec, latency *prometheus.HistogramVec, durationSummary *prometheus.SummaryVec) endpoint.ClientMiddleware {
	return func(c endpoint.HTTPClient) endpoint.HTTPClient {
		return func(r *http.Request) (resp *http.Response, err error) {
			defer func(start time.Time) {
				d := float64(time.Since(start).Nanoseconds())
				latency.WithLabelValues(serviceName).Observe(d)
				durationSummary.WithLabelValues(serviceName).Observe(d)
				requests.WithLabelValues(serviceName).Add(1)
			}(time.Now())
			return c.Do(r)
		}
	}
}
