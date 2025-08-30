package repository

import (
	"context"

	"ticketing/entity"

	"gorm.io/gorm"
)

type EventImageRepository interface {
	CreateImage(ctx context.Context, img *entity.EventImages) error
	GetByEventID(ctx context.Context, eventID uint) ([]entity.EventImages, error)
	GetByImageID(ctx context.Context, id uint) (*entity.EventImages, error)
	GetImageByIDIncludeDeleted(ctx context.Context, id uint) (*entity.EventImages, error)
	CountByEventID(ctx context.Context, eventID uint) (int64, error)
	UnsetPrimary(ctx context.Context, eventID uint) error
	DeleteImage(ctx context.Context, id uint) error
	RecoverImage(ctx context.Context, id uint) error
}

type eventImageRepo struct {
	db *gorm.DB
}

func NewEventImageRepository(db *gorm.DB) EventImageRepository {
	return &eventImageRepo{db: db}
}

func (r *eventImageRepo) CreateImage(ctx context.Context, img *entity.EventImages) error {
	return r.db.WithContext(ctx).Create(img).Error
}

func (r *eventImageRepo) GetByEventID(ctx context.Context, eventID uint) ([]entity.EventImages, error) {
	var imgs []entity.EventImages
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&imgs).Error
	return imgs, err
}

func (r *eventImageRepo) GetByImageID(ctx context.Context, id uint) (*entity.EventImages, error) {
	var img entity.EventImages
	err := r.db.WithContext(ctx).First(&img, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &img, nil
}

func (r *eventImageRepo) GetImageByIDIncludeDeleted(ctx context.Context, id uint) (*entity.EventImages, error) {
	var img entity.EventImages
	err := r.db.WithContext(ctx).Unscoped().First(&img, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &img, nil
}

func (r *eventImageRepo) CountByEventID(ctx context.Context, eventID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.EventImages{}).
		Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}

func (r *eventImageRepo) UnsetPrimary(ctx context.Context, eventID uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.EventImages{}).
		Where("event_id = ?", eventID).
		Update("is_primary", false).Error
}

func (r *eventImageRepo) DeleteImage(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.EventImages{}, id).Error
}

func (r *eventImageRepo) RecoverImage(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.EventImages{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}
