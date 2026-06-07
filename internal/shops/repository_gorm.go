package shops

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/shophub-platform/shophub/internal/models"
)

// gormRepository je GORM/PostgreSQL implementacija Repository-ja.
type gormRepository struct{ db *gorm.DB }

// NewGormRepository kreira Repository nad datom GORM konekcijom.
func NewGormRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) Create(ctx context.Context, shop *models.Shop) error {
	return r.db.WithContext(ctx).Create(shop).Error
}

func (r *gormRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&shops).Error
	return shops, err
}

func (r *gormRepository) GetByID(ctx context.Context, ownerID, id uuid.UUID) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, ownerID).
		First(&shop).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *gormRepository) Update(ctx context.Context, shop *models.Shop) error {
	return r.db.WithContext(ctx).
		Model(shop).
		Select("availability", "wallet_addr", "database_type", "image", "name").
		Updates(shop).Error
}

func (r *gormRepository) Delete(ctx context.Context, ownerID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, ownerID).
		Delete(&models.Shop{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
