package main

import (
	"net/http"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
)

func (m *Module) registerEndpoints() {
	m.MountingPoints["build"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			pn := q.Get("pn")
			branch := q.Get("branch")
			bn := q.Get("bn")
			basedir := q.Get("basedir")

			if pn == "" {
				http.Error(w, "missing projects name", http.StatusOK)
				return
			}

			if branch == "" {
				branch = "master"
			}

			result := m.publish(pn, branch, bn, basedir)
			http.Error(w, result, http.StatusOK)
		}),
		middlewares.BasicAuth(m, "build"),
	)
}
