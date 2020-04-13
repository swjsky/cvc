package root

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"loreal.com/dit/endpoint"
	"loreal.com/dit/middlewares"
	"loreal.com/dit/utils"
)

//RegisterEndpoints for SMS verify service
func (m *Module) registerEndpoints() {

	//Static file server
	m.MountingPoints[""] = endpoint.DecorateServer(
		http.StripPrefix(m.Prefix, http.FileServer(http.Dir("./public"))),
		//http.FileServer(http.Dir("./public")),
		//middlewares.FileServerRootRedirect(m.Prefix),
		middlewares.ServerInstrumentation("static", endpoint.RequestCounter, endpoint.LatencyHistogram, endpoint.DurationsSummary),
	)

	//metrics for prometheus system
	m.MountingPoints["metrics"] = endpoint.DecorateServer(
		prometheus.Handler(),
		middlewares.BasicAuthMd5(m, "metrics"),
	)

	//health check service
	m.MountingPoints["health"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		}),
		middlewares.NoCache(),
	)

	//reload
	m.MountingPoints["reload"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			m.Reload()
			w.Write([]byte("OK"))
		}),
		middlewares.BasicAuthMd5(m, "rootreload"),
	)

	m.MountingPoints["shutdown"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			q := r.URL.Query()
			shutdownKey := q.Get("key")
			if !m.verifyShutdownKey(shutdownKey) {
				http.Error(w, m.setShutdownKey(utils.RandomString(16)), http.StatusOK)
				return
			}
			w.Write([]byte("Shuting down..."))
			go m.Shutdown()
		}),
		middlewares.BasicAuthMd5(m, "rootshutdown"),
	)

	m.MountingPoints["restart"] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Supported", http.StatusNotAcceptable)
				return
			}
			m.Restart()
			w.Write([]byte("OK"))
		}),
		middlewares.BasicAuthMd5(m, "rootrestart"),
	)

}
