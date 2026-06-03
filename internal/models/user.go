package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User je ShopHub nalog koji upravlja sajtovima prodavnica (FZ 1.1).
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string    `gorm:"not null" json:"displayName"`

	// RefreshTokenVersion se uvećava pri logout-u / promeni lozinke kako bi
	// se prethodno izdati refresh tokeni poništili.
	RefreshTokenVersion int `gorm:"not null;default:0" json:"-"`

	Shops []Shop `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"shops,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate dodeljuje UUID ako nije postavljen.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
