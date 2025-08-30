package repository

import (
	"context"
	"time"

	"ticketing/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, u *entity.Users) error
	FindByEmail(ctx context.Context, email string) (*entity.Users, error)
	FindByUsername(ctx context.Context, username string) (*entity.Users, error)
	Update(ctx context.Context, u *entity.Users) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *entity.Users) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *entity.Users) error {
	u.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(u).Error
}

