package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

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
	Update(ctx context.Context, id string, condition domain.ProductCondition, conditionNote string, status domain.DetectionStatus) error
	UpdateStatus(ctx context.Context, id string, status domain.DetectionStatus) error
}

type DamageRepository interface {
	CreateAll(ctx context.Context, damages []domain.DamageCreate) error
}

type DetectorInput struct {
	ImageID     string
	GCSName     string
	URL         string
	ContentType string
	Angle       domain.ImageAngle
}

type DamageDetectionResult struct {
	Condition     domain.ProductCondition
	ConditionNote string
	Damages       []domain.DamageCreate
}

type DamageItemNotification struct {
	ImageID     string
	ImageURL    string
	ImageAngle  domain.ImageAngle
	DamageType  domain.DamageType
	BboxX1      *int
	BboxY1      *int
	BboxX2      *int
	BboxY2      *int
	Description *string
}

type DamageDetectionNotification struct {
	Condition     domain.ProductCondition
	ConditionNote string
	Damages       []DamageItemNotification
}

type DamageDetectionClient interface {
	Detect(ctx context.Context, images []DetectorInput) (DamageDetectionResult, error)
}

type DetectionNotifier interface {
	NotifyDamageDetectionComplete(userID string, notif DamageDetectionNotification)
	NotifyDamageDetectionFailed(userID string)
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
	damageRepo  DamageRepository
	detectionClient DamageDetectionClient
	notifier    DetectionNotifier
}

func NewImageService(
	s StorageClient,
	ir ProductImageRepository,
	sr DamageDetectionSummaryRepository,
	dr DamageRepository,
	detectionClient DamageDetectionClient,
	notifier DetectionNotifier,
) *ImageService {
	return &ImageService{
		storage:     s,
		imageRepo:   ir,
		summaryRepo: sr,
		damageRepo:  dr,
		detectionClient: detectionClient,
		notifier:    notifier,
	}
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

	summary, err := s.summaryRepo.Create(ctx, domain.DamageDetectionSummary{
		UserID: userID,
		Status: domain.DetectionStatusProcessing,
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

	detectorInputs := make([]DetectorInput, len(uploads))
	for i, u := range uploads {
		detectorInputs[i] = DetectorInput{
			ImageID:     ids[i],
			GCSName:     gcsNames[i],
			URL:         urls[i],
			ContentType: u.ContentType,
			Angle:       u.Angle,
		}
	}

	go s.runDetection(summary.ID, userID, detectorInputs)

	return ids, nil
}

func (s *ImageService) runDetection(summaryID, userID string, inputs []DetectorInput) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := s.detectionClient.Detect(ctx, inputs)
	if err != nil {
		slog.Error("damage detection failed", "summaryID", summaryID, "error", err)
		s.markDetectionFailed(summaryID, userID)
		return
	}

	if err := s.damageRepo.CreateAll(ctx, result.Damages); err != nil {
		slog.Error("failed to insert damages", "summaryID", summaryID, "error", err)
		s.markDetectionFailed(summaryID, userID)
		return
	}

	if err := s.summaryRepo.Update(ctx, summaryID, result.Condition, result.ConditionNote, domain.DetectionStatusDone); err != nil {
		slog.Error("failed to update damage detection summary", "summaryID", summaryID, "error", err)
		s.markDetectionFailed(summaryID, userID)
		return
	}

	s.notifier.NotifyDamageDetectionComplete(userID, buildNotification(result, inputs))
}

func buildNotification(result DamageDetectionResult, inputs []DetectorInput) DamageDetectionNotification {
	inputByID := make(map[string]DetectorInput, len(inputs))
	for _, inp := range inputs {
		inputByID[inp.ImageID] = inp
	}
	damages := make([]DamageItemNotification, len(result.Damages))
	for i, d := range result.Damages {
		inp := inputByID[d.ImageID]
		damages[i] = DamageItemNotification{
			ImageID:     d.ImageID,
			ImageURL:    inp.URL,
			ImageAngle:  inp.Angle,
			DamageType:  d.DamageType,
			BboxX1:      d.BboxX1,
			BboxY1:      d.BboxY1,
			BboxX2:      d.BboxX2,
			BboxY2:      d.BboxY2,
			Description: d.Description,
		}
	}
	return DamageDetectionNotification{
		Condition:     result.Condition,
		ConditionNote: result.ConditionNote,
		Damages:       damages,
	}
}

func (s *ImageService) markDetectionFailed(summaryID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.summaryRepo.UpdateStatus(ctx, summaryID, domain.DetectionStatusFailed); err != nil {
		slog.Error("failed to mark detection as failed", "summaryID", summaryID, "error", err)
	}
	s.notifier.NotifyDamageDetectionFailed(userID)
}

func (s *ImageService) deleteGCSObjects(names []string) {
	for _, name := range names {
		_ = s.storage.Delete(context.Background(), name)
	}
}
