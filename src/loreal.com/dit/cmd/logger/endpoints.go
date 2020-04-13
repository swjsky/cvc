package main

import (
	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"

	"github.com/microcosm-cc/bluemonday"
)

// var seededRand *rand.Rand
var sanitizePolicy *bluemonday.Policy

func init() {
	// 	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	sanitizePolicy = bluemonday.UGCPolicy()
}

func (a *App) initEndpoints() {
	a.Endpoints = map[string]EndpointEntry{}
}

// getDefaultMiddlewares - middlewares installed by defaults
func (a *App) getDefaultMiddlewares(path string) []endpoint.ServerMiddleware {
	return []endpoint.ServerMiddleware{
		middlewares.NoCache(),
		middlewares.CORS("*", "*", "Content-Type, Accept, Authorization", ""),
		middlewares.BasicAuthOrTokenAuthWithRole(a.AuthProvider, "", "user,admin"),
		middlewares.ServerInstrumentation(path, endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	}
}
