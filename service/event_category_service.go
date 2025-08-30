package service

import (
	"context"
	"errors"
	"time"

	"ticketing/entity"
	"ticketing/repository"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, c *entity.EventCategories) error
	GetCategoryByIDIncludeDeleted(ctx context.Context, id uint) (*entity.EventCategories, error)
	GetCategories(ctx context.Context, limit, offset int) ([]entity.EventCategories, error)
	UpdateCategory(ctx context.Context, c *entity.EventCategories) error
	DeleteCategory(ctx context.Context, id uint) error
	RecoverCategory(ctx context.Context, id uint) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, c *entity.EventCategories) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return errors.New("failed to create category: " + err.Error())
	}
	return nil
}

func (s *categoryService) GetCategoryByIDIncludeDeleted(ctx context.Context, id uint) (*entity.EventCategories, error) {
	cat, err := s.repo.GetCategoryByIDIncludeDeleted(ctx, id)
	if err != nil {
		return nil, errors.New("failed to fetch category: " + err.Error())
	}
	if cat == nil {
		return nil, errors.New("category not found")
	}
	return cat, nil
}

func (s *categoryService) GetCategories(ctx context.Context, limit, offset int) ([]entity.EventCategories, error) {
	categories, err := s.repo.GetCategories(ctx, limit, offset)
	if err != nil {
		return nil, errors.New("failed to fetch categories: " + err.Error())
	}
	return categories, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, c *entity.EventCategories) error {
	existing, err := s.repo.GetCategoryByIDIncludeDeleted(ctx, c.ID)
	if err != nil {
		return errors.New("failed to fetch category: " + err.Error())
	}
	if existing == nil {
		return errors.New("category not found")
	}

	// Update fields
	existing.Name = c.Name
	if c.Events != nil {
		existing.Events = c.Events
	}

	if err := s.repo.UpdateCategory(ctx, existing); err != nil {
		return errors.New("failed to update category: " + err.Error())
	}
	return nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id uint) error {
	existing, err := s.repo.GetCategoryByIDIncludeDeleted(ctx, id)
	if err != nil {
		return errors.New("failed to fetch category: " + err.Error())
	}
	if existing == nil {
		return errors.New("category not found")
	}
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return errors.New("failed to delete category: " + err.Error())
	}
	return nil
}

func (s *categoryService) RecoverCategory(ctx context.Context, id uint) error {
	existing, err := s.repo.GetCategoryByIDIncludeDeleted(ctx, id)
	if err != nil {
		return errors.New("failed to fetch category: " + err.Error())
	}
	if existing == nil {
		return errors.New("category not found")
	}
	if err := s.repo.RecoverCategory(ctx, id); err != nil {
		return errors.New("failed to recover category: " + err.Error())
	}
	return nil
}
