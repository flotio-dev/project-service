package api

import (
	"net/http"

	"github.com/flotio-dev/project-service/pkg/db"
	"github.com/flotio-dev/project-service/pkg/httpx"
	"github.com/flotio-dev/project-service/pkg/middleware"
	"github.com/gorilla/mux"
)

func (a *API) mountEnvVars(api *mux.Router) {
	api.HandleFunc("/envvars", func(w http.ResponseWriter, r *http.Request) {
		sub, groups := getUserAndGroups(r)
		var envs []db.EnvVar
		q := a.DB.Model(&db.EnvVar{}).Joins("JOIN projects ON projects.id = env_vars.project_id")
		if len(groups) > 0 {
			q = q.Where("projects.user_id = ? OR projects.group_id IN ?", sub, groups)
		} else {
			q = q.Where("projects.user_id = ?", sub)
		}
		if err := q.Order("env_vars.created_at DESC").Find(&envs).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.OK(w, envs)
	}).Methods(http.MethodGet)
}

// helper partagé avec projects.go
func getUserAndGroups(r *http.Request) (string, []string) {
	sub, _ := middleware.GetValue[string](r, "sub")
	var groups []string
	if claims, ok := middleware.GetValue[map[string]any](r, "claims"); ok {
		if raw, ok := claims["groups"]; ok {
			switch v := raw.(type) {
			case []any:
				for _, it := range v {
					if s, _ := it.(string); s != "" {
						groups = append(groups, s)
					}
				}
			case []string:
				groups = append(groups, v...)
			}
		}
	}
	return sub, groups
}
