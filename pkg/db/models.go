package db

import (
	"time"
)

// Project représente un projet utilisateur
type Project struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	UserID       string  `gorm:"index;not null" json:"user_id"`   // Keycloak sub
	GroupID      *string `gorm:"index" json:"group_id,omitempty"` // Groupe optionnel
	Name         string  `gorm:"not null" json:"name"`
	GithubToken  *string `json:"github_token,omitempty"`
	GithubRepo   *string `gorm:"index" json:"github_repo,omitempty"` // owner/repo
	GithubURL    *string `json:"github_url,omitempty"`               // https://github.com/owner/repo
	Subscription *int    `json:"subscription_used,omitempty"`        // nombre d'abonnements utilisés

	Stats   []ProjectStats `gorm:"foreignKey:ProjectID" json:"-"`
	Usages  []ProjectUsage `gorm:"foreignKey:ProjectID" json:"-"`
	Branches []Branch      `gorm:"foreignKey:ProjectID" json:"-"`
	EnvVars []EnvVar       `gorm:"foreignKey:ProjectID" json:"-"`
	Builds  []Build        `gorm:"foreignKey:ProjectID" json:"-"`
}

// ProjectStats stocke des statistiques journalières
type ProjectStats struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectID   string    `gorm:"type:uuid;index;not null" json:"project_id"`
	Day         time.Time `gorm:"type:date;index" json:"day"`
	UniqueOpens int       `gorm:"not null;default:0" json:"unique_opens"`
	Opens       int       `gorm:"not null;default:0" json:"opens"`
}

// ProjectUsage recense le nombre de builds par plateforme et par mois
type ProjectUsage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectID string `gorm:"type:uuid;index;not null" json:"project_id"`
	Year      int    `gorm:"index" json:"year"`
	Month     int    `gorm:"index" json:"month"`            // 1-12
	Platform  string `gorm:"index;size:16" json:"platform"` // IOS, ANDROID, LINUX, WINDOWS, MAC
	Builds    int    `gorm:"not null;default:0" json:"builds"`
}

// Branch représente une branche d'un projet
type Branch struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	ProjectID string `gorm:"type:uuid;index;not null" json:"project_id"`
	Name      string `gorm:"not null" json:"name"`
}

// EnvVar représente une variable d'environnement (texte ou fichier)
type EnvVar struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	ProjectID string `gorm:"type:uuid;index;not null" json:"project_id"`
	Key       string `gorm:"index;not null" json:"key"`
	Category  string `gorm:"index;size:64" json:"category"` // ex: push-notification, firebase-config, etc.
	// Type: "text" ou "file"
	Type string `gorm:"size:8;not null" json:"type"`
	// Si Type==text, la valeur est stockée ici
	Value *string `json:"value,omitempty"`
	// Si Type==file, on stocke une URL vers le fichier
	FileURL *string `json:"file_url,omitempty"`
}

// Build représente un build d'un projet, associé éventuellement à une branche et une plateforme
type Build struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	ProjectID string  `gorm:"type:uuid;index;not null" json:"project_id"`
	BranchID  *string `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Platform  string  `gorm:"index;size:16" json:"platform"`                          // IOS, ANDROID, LINUX, WINDOWS, MAC
	Status    string  `gorm:"index;size:16;not null;default:'pending'" json:"status"` // pending, running, success, failed

	DownloadURL string `gorm:"not null" json:"download_url"`

	Logs []BuildLog `gorm:"foreignKey:BuildID" json:"-"`
}

// BuildLog stocke des lignes de log séquentielles pour un build
type BuildLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	// Pas d'UpdatedAt nécessaire pour des logs append-only

	BuildID string `gorm:"type:uuid;index;not null" json:"build_id"`
	// Un ordre croissant pour rejouer l'historique facilement
	Seq  int    `gorm:"not null;index" json:"seq"`
	Line string `gorm:"type:text;not null" json:"line"`
}
