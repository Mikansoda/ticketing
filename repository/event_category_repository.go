package repository

import (
	"context"

	"ticketing/entity"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, c *entity.EventCategories) error
	GetCategories(ctx context.Context, limit, offset int) ([]entity.EventCategories, error)
	GetCategoryByIDIncludeDeleted(ctx context.Context, id uint) (*entity.EventCategories, error)
	UpdateCategory(ctx context.Context, c *entity.EventCategories) error
	DeleteCategory(ctx context.Context, id uint) error
	RecoverCategory(ctx context.Context, id uint) error
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) CreateCategory(ctx context.Context, c *entity.EventCategories) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *categoryRepo) GetCategories(ctx context.Context, limit, offset int) ([]entity.EventCategories, error) {
	var categories []entity.EventCategories
	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepo) GetCategoryByIDIncludeDeleted(ctx context.Context, id uint) (*entity.EventCategories, error) {
	var c entity.EventCategories
	if err := r.db.WithContext(ctx).
		Unscoped().
		First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) UpdateCategory(ctx context.Context, c *entity.EventCategories) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *categoryRepo) DeleteCategory(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.EventCategories{}, id).Error
}

func (r *categoryRepo) RecoverCategory(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.EventCategories{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}
