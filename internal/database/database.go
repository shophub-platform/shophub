// Package database otvara GORM konekciju i pokreće migracije.
package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/shophub-platform/shophub/internal/models"
)

// Connect otvara GORM konekciju ka PostgreSQL bazi.
func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrate pokreće GORM auto-migracije za sve modele.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Shop{})
}
