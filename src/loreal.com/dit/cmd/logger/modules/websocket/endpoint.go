package websocket

import (
	"log"
	"net/http"

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

func (m *Module) registerEndpoints() {
	m.MountingPoints[""] = endpoint.DecorateServer(
		endpoint.Impl(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
				return
			}
			q := r.URL.Query()
			projectID := sanitizePolicy.Sanitize(q.Get("project_id"))
			if projectID == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			upgrader.CheckOrigin = CheckOriginURL
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "error params", http.StatusBadRequest)
				return
			}
			websocket := NewWebsocket(projectID, conn)
			websocket.Listen()
		}),
		middlewares.NoCache())
}
