package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	gcfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	return gorm.Open(postgres.Open(databaseURL), gcfg)
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Project{},
		&ProjectStats{},
		&ProjectUsage{},
		&Branch{},
		&EnvVar{},
		&Build{},
		&BuildLog{},
	)
}

func Must(db *gorm.DB, err error) *gorm.DB {
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	return db
}
