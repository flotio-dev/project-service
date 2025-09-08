package api

import (
	"github.com/flotio-dev/project-service/pkg/auth"
	"gorm.io/gorm"
)

// API regroupe les dépendances nécessaires aux handlers
type API struct {
	DB   *gorm.DB
	JWKS *auth.JWKSProvider
}
