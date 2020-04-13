package main

import (
	"net/http"

	"loreal.com/dit/module/modules/root"
	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
)

func registerEndpoints(m *root.Module, u middlewares.RoleVerifier) {

	m.MountingPoints[""] = endpoint.DecorateServer(
		http.StripPrefix(m.Prefix, http.FileServer(http.Dir(Cfg.ShareFolder))),
		middlewares.ServerInstrumentation("files", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
		middlewares.BasicAuthOrTokenAuthWithRoleSkipable(u, "", "user,admin", !Cfg.Auth),
		middlewares.NoCache(),
	)

}
