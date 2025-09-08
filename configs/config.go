package configs

import (
	"fmt"
	"os"
	"strconv"
)

// Config regroupe la configuration de l'application.
type Config struct {
	// HTTP
	HTTPPort int

	// Base de données
	DatabaseURL string

	// Keycloak / OpenID Connect
	KeycloakBaseURL string // ex: https://auth.example.com
	KeycloakRealm   string // ex: my-realm
}

// JWKSURL retourne l'URL JWKS de Keycloak.
func (c Config) JWKSURL() string {
	if c.KeycloakBaseURL == "" || c.KeycloakRealm == "" {
		return ""
	}
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.KeycloakBaseURL, c.KeycloakRealm)
}

// IssuerURL retourne l'issuer attendu pour la validation des tokens.
func (c Config) IssuerURL() string {
	if c.KeycloakBaseURL == "" || c.KeycloakRealm == "" {
		return ""
	}
	return fmt.Sprintf("%s/realms/%s", c.KeycloakBaseURL, c.KeycloakRealm)
}

// FromEnv charge la configuration depuis les variables d'environnement.
func FromEnv() (Config, error) {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	return Config{
		HTTPPort:        port,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		KeycloakBaseURL: os.Getenv("KEYCLOAK_BASE_URL"),
		KeycloakRealm:   os.Getenv("KEYCLOAK_REALM"),
	}, nil
}
