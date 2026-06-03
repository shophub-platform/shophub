package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Availability odgovara dostupnosti aplikacije iz FZ 1.2 i CRD-a operatora.
type Availability string

const (
	AvailabilityStandard Availability = "standard" // 2 replike
	AvailabilityHigh     Availability = "high"      // 3 replike
)

// DatabaseType odgovara izboru baze iz FZ 1.2.
type DatabaseType string

const (
	DatabasePostgres DatabaseType = "postgres"
	DatabaseRedis    DatabaseType = "redis"
)

// Shop je sajt prodavnice koji pripada korisniku. Mapira se na Shop CR
// koji reconcile-uje shop-operator.
type Shop struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID      uuid.UUID    `gorm:"type:uuid;index;not null" json:"ownerId"`
	Name         string       `gorm:"not null" json:"name"`
	Availability Availability `gorm:"type:varchar(16);not null;default:standard" json:"availability"`
	WalletAddr   string       `gorm:"not null" json:"walletAddress"`
	DatabaseType DatabaseType `gorm:"type:varchar(16);not null;default:postgres" json:"databaseType"`
	Image        string       `gorm:"not null" json:"image"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Replicas vraća broj replika koji proizlazi iz dostupnosti.
func (s Shop) Replicas() int {
	if s.Availability == AvailabilityHigh {
		return 3
	}
	return 2
}

// BeforeCreate dodeljuje UUID ako nije postavljen.
func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
