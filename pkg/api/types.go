package api

import (
	"github.com/flotio-dev/user-service/pkg/auth"
	"gorm.io/gorm"
)

// API regroupe les dépendances nécessaires aux handlers
type API struct {
	DB   *gorm.DB
	JWKS *auth.JWKSProvider
}
