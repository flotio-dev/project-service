package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/flotio-dev/project-service/configs"
	"github.com/flotio-dev/project-service/pkg/api"
	"github.com/flotio-dev/project-service/pkg/auth"
	"github.com/flotio-dev/project-service/pkg/db"
)

func main() {
	cfg, _ := configs.FromEnv()
	if cfg.DatabaseURL == "" {
		log.Println("warning: DATABASE_URL is empty")
	}
	log.Printf("connecting to database: %s", cfg.DatabaseURL)
	gdb := db.Must(db.Connect(cfg.DatabaseURL))
	log.Println("database connected")

	log.Println("running automigrate...")
	if err := db.AutoMigrate(gdb); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}
	log.Println("automigrate completed")

	// JWKS provider pour Keycloak
	jwksURL := cfg.JWKSURL()
	issuer := cfg.IssuerURL()
	var jwksProv *auth.JWKSProvider
	if jwksURL != "" {
		jwksProv = auth.NewJWKSProvider(jwksURL, issuer)
	}

	apiSrv := &api.API{DB: gdb, JWKS: jwksProv}
	r := apiSrv.Router()
	log.Println("router constructed")

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
