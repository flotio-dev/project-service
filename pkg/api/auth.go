package api

import (
	"net/http"

	"github.com/flotio-dev/project-service/pkg/httpx"
	"github.com/flotio-dev/project-service/pkg/middleware"
	"github.com/gorilla/mux"
)

func (a *API) mountAuth(api *mux.Router) {
	api.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		sub, _ := middleware.GetValue[string](r, "sub")
		httpx.OK(w, map[string]any{"sub": sub})
	}).Methods(http.MethodGet)
}
