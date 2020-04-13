package endpoint

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//DefaultHTTPClient - default HTTP Client
var DefaultHTTPClient *http.Client

//RequestCounter - default RequestsCounter
var RequestCounter *prometheus.CounterVec

//LatencyHistogram - default Latency Histogram
var LatencyHistogram *prometheus.HistogramVec

//DurationsSummary - default Latency Summary
var DurationsSummary *prometheus.SummaryVec

func init() {
	SetPrometheus("BaseSystem")

}

//SetPrometheus - set subsystem for prometheus
func SetPrometheus(subsystem string) {
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "CEH",
			Subsystem: subsystem,
			Name:      "RequestCounter",
			Help:      "Requests Count",
		},
		[]string{"service"},
	)

	LatencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "CEH",
			Subsystem: subsystem,
			Name:      "Latency",
			Help:      "Requests Latency Histogram",
			Buckets:   []float64{0, 1e+5 /*100 us*/, 1e+6 /*1 ms*/, 1e+7 /*10 ms*/, 1e+8 /*100 ms*/, 1e+9 /*1 s*/, 1e+10 /*10 s*/},
		},
		[]string{"service"},
	)

	DurationsSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "CEH",
			Subsystem: subsystem,
			Name:      "LatencySummary",
			Help:      "Requests Latency Summary",
		},
		[]string{"service"},
	)

	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(LatencyHistogram)
	prometheus.MustRegister(DurationsSummary)

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		Proxy:                 nil,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          50,
		IdleConnTimeout:       30 * time.Second,
	}
	DefaultHTTPClient = &http.Client{Transport: tr}
}
