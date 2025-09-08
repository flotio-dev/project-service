package api

import (
	"encoding/json"
	"net/http"

	"github.com/flotio-dev/project-service/pkg/db"
	"github.com/flotio-dev/project-service/pkg/httpx"
	"github.com/flotio-dev/project-service/pkg/middleware"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func (a *API) mountBuilds(api *mux.Router) {
	// POST /api/projects/{projectID}/builds
	api.HandleFunc("/projects/{projectID}/builds", func(w http.ResponseWriter, r *http.Request) {
		sub, _ := middleware.GetValue[string](r, "sub")
		vars := mux.Vars(r)
		projectID := vars["projectID"]

		var p db.Project
		if err := a.DB.First(&p, "id = ? AND user_id = ?", projectID, sub).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				httpx.NotFound(w, "project not found")
				return
			}
			httpx.InternalError(w, err.Error())
			return
		}

		var in struct {
			BranchID    *string `json:"branch_id"`
			Platform    string  `json:"platform"`
			DownloadURL string  `json:"download_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			httpx.BadRequest(w, "invalid json")
			return
		}
		if in.Platform == "" || in.DownloadURL == "" {
			httpx.BadRequest(w, "platform and download_url required")
			return
		}
		b := db.Build{ProjectID: projectID, BranchID: in.BranchID, Platform: in.Platform, DownloadURL: in.DownloadURL, Status: "success"}
		if err := a.DB.Create(&b).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.Created(w, b)
	}).Methods(http.MethodPost)

	// GET /api/projects/{projectID}/builds
	api.HandleFunc("/projects/{projectID}/builds", func(w http.ResponseWriter, r *http.Request) {
		sub, _ := middleware.GetValue[string](r, "sub")
		vars := mux.Vars(r)
		projectID := vars["projectID"]

		var count int64
		if err := a.DB.Model(&db.Project{}).Where("id = ? AND user_id = ?", projectID, sub).Count(&count).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		if count == 0 {
			httpx.NotFound(w, "project not found")
			return
		}
		var builds []db.Build
		if err := a.DB.Where("project_id = ?", projectID).Order("created_at DESC").Find(&builds).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.OK(w, builds)
	}).Methods(http.MethodGet)

	// GET /api/builds/{buildID}/logs
	api.HandleFunc("/builds/{buildID}/logs", func(w http.ResponseWriter, r *http.Request) {
		sub, _ := middleware.GetValue[string](r, "sub")
		buildID := mux.Vars(r)["buildID"]

		var b db.Build
		if err := a.DB.First(&b, "id = ?", buildID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				httpx.NotFound(w, "build not found")
				return
			}
			httpx.InternalError(w, err.Error())
			return
		}
		var p db.Project
		if err := a.DB.Select("id,user_id").First(&p, "id = ?", b.ProjectID).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		if p.UserID != sub {
			httpx.Forbidden(w, "forbidden")
			return
		}
		var logs []db.BuildLog
		if err := a.DB.Where("build_id = ?", buildID).Order("seq ASC").Find(&logs).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.OK(w, logs)
	}).Methods(http.MethodGet)
}
