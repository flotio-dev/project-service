package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flotio-dev/project-service/pkg/db"
	"github.com/flotio-dev/project-service/pkg/httpx"
	"github.com/flotio-dev/project-service/pkg/middleware"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func (a *API) mountProjects(api *mux.Router) {
	// helpers
	type githubRepo struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	}
	getUserAndGroups := func(r *http.Request) (string, []string) {
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
	hasAccess := func(p db.Project, sub string, groups []string) bool {
		if p.UserID == sub {
			return true
		}
		if p.GroupID != nil {
			gid := *p.GroupID
			for _, g := range groups {
				if g == gid {
					return true
				}
			}
		}
		return false
	}

	// Create project
	api.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		sub, _ := middleware.GetValue[string](r, "sub")
		var in struct {
			Name        string  `json:"name"`
			GroupID     *string `json:"group_id"`
			GithubToken *string `json:"github_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
			httpx.BadRequest(w, "invalid payload (name required)")
			return
		}
		p := db.Project{UserID: sub, GroupID: in.GroupID, Name: in.Name, GithubToken: in.GithubToken}
		if err := a.DB.Create(&p).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.Created(w, p)
	}).Methods(http.MethodPost)

	// List projects
	api.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		sub, groups := getUserAndGroups(r)
		var ps []db.Project
		q := a.DB.Model(&db.Project{})
		if len(groups) > 0 {
			q = q.Where("user_id = ? OR group_id IN ?", sub, groups)
		} else {
			q = q.Where("user_id = ?", sub)
		}
		if err := q.Order("created_at DESC").Find(&ps).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.OK(w, ps)
	}).Methods(http.MethodGet)

	// Get one project
	api.HandleFunc("/projects/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		sub, groups := getUserAndGroups(r)
		id := mux.Vars(r)["projectID"]
		var p db.Project
		if err := a.DB.First(&p, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httpx.NotFound(w, "project not found")
				return
			}
			httpx.InternalError(w, err.Error())
			return
		}
		if !hasAccess(p, sub, groups) {
			httpx.Forbidden(w, "forbidden")
			return
		}
		httpx.OK(w, p)
	}).Methods(http.MethodGet)

	// Update project
	api.HandleFunc("/projects/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		sub, groups := getUserAndGroups(r)
		id := mux.Vars(r)["projectID"]
		var p db.Project
		if err := a.DB.First(&p, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httpx.NotFound(w, "project not found")
				return
			}
			httpx.InternalError(w, err.Error())
			return
		}
		if !hasAccess(p, sub, groups) {
			httpx.Forbidden(w, "forbidden")
			return
		}
		var in struct {
			Name        *string `json:"name"`
			GroupID     *string `json:"group_id"`
			GithubToken *string `json:"github_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			httpx.BadRequest(w, "invalid payload")
			return
		}
		updates := map[string]any{}
		if in.Name != nil {
			updates["name"] = *in.Name
		}
		if in.GroupID != nil {
			updates["group_id"] = in.GroupID
		}
		if in.GithubToken != nil {
			updates["github_token"] = in.GithubToken
		}
		if len(updates) == 0 {
			httpx.OK(w, p)
			return
		}
		if err := a.DB.Model(&p).Updates(updates).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.OK(w, p)
	}).Methods(http.MethodPatch, http.MethodPut)

	// Delete project
	api.HandleFunc("/projects/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		sub, groups := getUserAndGroups(r)
		id := mux.Vars(r)["projectID"]
		var p db.Project
		if err := a.DB.First(&p, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httpx.NotFound(w, "project not found")
				return
			}
			httpx.InternalError(w, err.Error())
			return
		}
		if !hasAccess(p, sub, groups) {
			httpx.Forbidden(w, "forbidden")
			return
		}
		if err := a.DB.Delete(&p).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.NoContent(w)
	}).Methods(http.MethodDelete)

	// Import GitHub
	api.HandleFunc("/projects/import/github", func(w http.ResponseWriter, r *http.Request) {
		sub, _ := middleware.GetValue[string](r, "sub")
		var in struct {
			Token    string  `json:"token"`
			FullName *string `json:"full_name"`
			Query    *string `json:"query"`
			GroupID  *string `json:"group_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Token == "" {
			httpx.BadRequest(w, "invalid payload (token required)")
			return
		}
		var repo githubRepo
		client := &http.Client{Timeout: 10 * time.Second}
		if in.FullName != nil && *in.FullName != "" {
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s", *in.FullName), nil)
			req.Header.Set("Authorization", "Bearer "+in.Token)
			req.Header.Set("Accept", "application/vnd.github+json")
			res, err := client.Do(req)
			if err != nil || res.StatusCode != 200 {
				httpx.BadRequest(w, "cannot fetch repo; check full_name/token")
				return
			}
			defer res.Body.Close()
			if err := json.NewDecoder(res.Body).Decode(&repo); err != nil {
				httpx.InternalError(w, err.Error())
				return
			}
		} else if in.Query != nil && *in.Query != "" {
			type searchResp struct {
				Items []githubRepo `json:"items"`
			}
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/search/repositories?q=%s&per_page=5", *in.Query), nil)
			req.Header.Set("Authorization", "Bearer "+in.Token)
			req.Header.Set("Accept", "application/vnd.github+json")
			res, err := client.Do(req)
			if err != nil || res.StatusCode != 200 {
				httpx.BadRequest(w, "github search failed")
				return
			}
			defer res.Body.Close()
			var sr searchResp
			if err := json.NewDecoder(res.Body).Decode(&sr); err != nil {
				httpx.InternalError(w, err.Error())
				return
			}
			if len(sr.Items) == 0 {
				httpx.NotFound(w, "no repository matched query")
				return
			}
			repo = sr.Items[0]
		} else {
			httpx.BadRequest(w, "provide full_name or query")
			return
		}
		name := repo.Name
		p := db.Project{UserID: sub, GroupID: in.GroupID, Name: name, GithubToken: &in.Token}
		if repo.FullName != "" {
			p.GithubRepo = &repo.FullName
		}
		if repo.HTMLURL != "" {
			p.GithubURL = &repo.HTMLURL
		}
		if err := a.DB.Create(&p).Error; err != nil {
			httpx.InternalError(w, err.Error())
			return
		}
		httpx.Created(w, p)
	}).Methods(http.MethodPost)
}
