package service

import (
	"context"
	"fmt"
	"io"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/google/uuid"
)

type StorageClient interface {
	Upload(ctx context.Context, name string, r io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, name string) error
}

type ProductImageRepository interface {
	CreateAll(ctx context.Context, images []domain.ProductImage) ([]string, error)
}

type DamageDetectionSummaryRepository interface {
	Create(ctx context.Context, summary domain.DamageDetectionSummary) (domain.DamageDetectionSummary, error)
}

type ImageUpload struct {
	Reader      io.Reader
	ContentType string
	Angle       domain.ImageAngle
}

type ImageService struct {
	storage     StorageClient
	imageRepo   ProductImageRepository
	summaryRepo DamageDetectionSummaryRepository
}

func NewImageService(s StorageClient, ir ProductImageRepository, sr DamageDetectionSummaryRepository) *ImageService {
	return &ImageService{storage: s, imageRepo: ir, summaryRepo: sr}
}

func (s *ImageService) UploadImages(ctx context.Context, userID string, uploads []ImageUpload) ([]string, error) {
	gcsNames := make([]string, 0, len(uploads))
	urls := make([]string, len(uploads))
	for i, u := range uploads {
		ext := ".jpg"
		if u.ContentType == "image/png" {
			ext = ".png"
		}
		name := fmt.Sprintf("product-images/%s%s", uuid.New().String(), ext)
		url, err := s.storage.Upload(ctx, name, u.Reader, u.ContentType)
		if err != nil {
			s.deleteGCSObjects(gcsNames)
			return nil, apperror.ErrInternal.Wrap(err, "failed to upload image to GCS")
		}
		gcsNames = append(gcsNames, name)
		urls[i] = url
	}

	// stub: #13 でurlsをVertex AIに渡して傷検出を呼び出し、結果をCondition/ConditionNoteに反映する
	summary, err := s.summaryRepo.Create(ctx, domain.DamageDetectionSummary{
		UserID:        userID,
		Condition:     domain.ConditionGood,
		ConditionNote: "",
	})
	if err != nil {
		s.deleteGCSObjects(gcsNames)
		return nil, err
	}

	images := make([]domain.ProductImage, len(uploads))
	for i, u := range uploads {
		images[i] = domain.ProductImage{
			SummaryID: &summary.ID,
			URL:       urls[i],
			Angle:     u.Angle,
		}
	}

	ids, err := s.imageRepo.CreateAll(ctx, images)
	if err != nil {
		s.deleteGCSObjects(gcsNames)
		return nil, err
	}
	return ids, nil
}

func (s *ImageService) deleteGCSObjects(names []string) {
	for _, name := range names {
		_ = s.storage.Delete(context.Background(), name)
	}
}
