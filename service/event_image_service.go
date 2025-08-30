package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"ticketing/config"
	"ticketing/entity"
	"ticketing/repository"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type EventImageService interface {
	Upload(ctx context.Context, eventID uint, filePath string, isPrimary bool) (*entity.EventImages, error)
	GetByEventID(ctx context.Context, eventID uint) ([]entity.EventImages, error)
	Delete(ctx context.Context, id uint) error
	Recover(ctx context.Context, id uint) error
}

type eventImageService struct {
	repo  repository.EventImageRepository
	cloud *cloudinary.Cloudinary
}

func NewEventImagesService(repo repository.EventImageRepository) EventImageService {
	cloud := config.InitCloud()
	return &eventImageService{
		repo:  repo,
		cloud: cloud,
	}
}

func (s *eventImageService) Upload(ctx context.Context, eventID uint, filePath string, isPrimary bool) (*entity.EventImages, error) {
	// max 3 images check
	count, err := s.repo.CountByEventID(ctx, eventID)
	if err != nil {
		return nil, errors.New("failed to count event images: " + err.Error())
	}
	if count >= 3 {
		return nil, errors.New("maximum 3 images per event")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("failed to open file: " + err.Error())
	}
	defer f.Close()

	uploadRes, err := s.cloud.Upload.Upload(ctx, f, uploader.UploadParams{
		Folder: "ticketing/events",
	})
	if err != nil {
		return nil, errors.New("failed to upload image: " + err.Error())
	}

	fmt.Printf("DEBUG uploadRes: %+v\n", uploadRes)

	if uploadRes == nil || uploadRes.SecureURL == "" {
		return nil, errors.New("failed to upload image: empty secure URL from cloud")
	}

	if isPrimary {
		if err := s.repo.UnsetPrimary(ctx, eventID); err != nil {
			return nil, errors.New("failed to unset previous primary image: " + err.Error())
		}
	}

	img := &entity.EventImages{
		EventID:   eventID,
		ImageURL:  uploadRes.SecureURL,
		IsPrimary: isPrimary,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateImage(ctx, img); err != nil {
		return nil, errors.New("failed to save event image: " + err.Error())
	}
	return img, nil
}

func (s *eventImageService) GetByEventID(ctx context.Context, eventID uint) ([]entity.EventImages, error) {
	images, err := s.repo.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, errors.New("failed to fetch event images: " + err.Error())
	}
	return images, nil
}

func (s *eventImageService) Delete(ctx context.Context, id uint) error {
	existing, err := s.repo.GetByImageID(ctx, id)
	if err != nil {
		return errors.New("failed to fetch event image: " + err.Error())
	}
	if existing == nil {
		return errors.New("event image not found")
	}
	if err := s.repo.DeleteImage(ctx, id); err != nil {
		return errors.New("failed to delete event image: " + err.Error())
	}
	return nil
}

func (s *eventImageService) Recover(ctx context.Context, id uint) error {
	existing, err := s.repo.GetImageByIDIncludeDeleted(ctx, id)
	if err != nil {
		return errors.New("failed to fetch event image: " + err.Error())
	}
	if existing == nil {
		return errors.New("event image not found")
	}
	if err := s.repo.RecoverImage(ctx, id); err != nil {
		return errors.New("failed to recover event image: " + err.Error())
	}
	return nil
}
