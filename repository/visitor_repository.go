package repository

import (
	"context"
	"ticketing/entity"

	"gorm.io/gorm"
)

type VisitorRepository interface {
	GetByNationalID(ctx context.Context, nid string) (*entity.Visitors, error)
	Create(ctx context.Context, v *entity.Visitors, tx *gorm.DB) error
}

type visitorRepo struct {
	db *gorm.DB
}

func NewVisitorRepo(db *gorm.DB) VisitorRepository {
	return &visitorRepo{db}
}

func (r *visitorRepo) GetByNationalID(ctx context.Context, nid string) (*entity.Visitors, error) {
	var v entity.Visitors
	if err := r.db.WithContext(ctx).Where("national_id = ?", nid).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *visitorRepo) Create(ctx context.Context, v *entity.Visitors, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(v).Error
}
