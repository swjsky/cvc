package main

import (
	"net"
	"net/http"
	"time"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
)

var httpClient endpoint.HTTPClient

func init() {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		Proxy:                 nil,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          50,
		IdleConnTimeout:       30 * time.Second,
	}
	endpoint.DefaultHTTPClient = &http.Client{Transport: tr}

	httpClient = endpoint.DecorateClient(endpoint.DefaultHTTPClient,
		middlewares.FaultTolerance(3, 5*time.Second, nil),
		//middlewares.ClientInstrumentation("wx-msg-backend", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	)
}
