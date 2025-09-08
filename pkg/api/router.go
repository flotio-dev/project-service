package api

import (
	"net/http"
	"time"

	"github.com/flotio-dev/project-service/pkg/httpx"
	"github.com/flotio-dev/project-service/pkg/middleware"
	"github.com/gorilla/mux"
)

func (a *API) Router() http.Handler {
	r := mux.NewRouter()
	// global logging middleware
	r.Use(middleware.LoggingMiddleware)
	// Public
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]any{"status": "ok", "time": time.Now()})
	}).Methods(http.MethodGet)

	// Protected API
	api := r.PathPrefix("/api").Subrouter()
	if a.JWKS != nil {
		api.Use(middleware.RequireAuth(a.JWKS, ""))
	}

	// Mount per-model subrouters
	a.mountProjects(api)
	a.mountBuilds(api)
	a.mountEnvVars(api)
	a.mountAuth(api)
	return r
}
