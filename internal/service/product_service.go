package service

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error)
	ListBySeller(ctx context.Context, sellerID string, f domain.ListingsFilter) ([]domain.Product, int, error)
	GetByID(ctx context.Context, id string, uid *string) (domain.ProductDetail, error)
	Create(ctx context.Context, sellerID string, input domain.ProductCreate) (domain.Product, error)
	Update(ctx context.Context, id string, sellerID string, input domain.ProductUpdate) error
	Delete(ctx context.Context, id string, sellerID string) error
}

type ViewingHistoryRepository interface {
	Upsert(ctx context.Context, userID, productID string) error
	ListByUserID(ctx context.Context, userID string, f domain.ViewingHistoryFilter) ([]domain.ViewingHistory, int, error)
}

type ProductModelRepository interface {
	Create(ctx context.Context, productID string) (string, error)
	UpdateJobID(ctx context.Context, id, jobID string) error
	UpdateStatus(ctx context.Context, id string, status domain.ModelStatus) error
	UpdateDone(ctx context.Context, id, glbURL string) error
}

type ProductImageURLRepository interface {
	GetURLsByProductID(ctx context.Context, productID string) ([]string, error)
}

type ModelGenerationClient interface {
	CreateJob(ctx context.Context, imageURLs []string) (string, error)
	GetJobStatus(ctx context.Context, jobID string) (status string, glbURL string, err error)
}

type ModelGenerationNotification struct {
	ProductID string
	GlbURL    string
}

type ModelNotifier interface {
	NotifyModelGenerationComplete(userID string, notif ModelGenerationNotification)
	NotifyModelGenerationFailed(userID, productID string)
}

type ProductService struct {
	repo          ProductRepository
	historyRepo   ViewingHistoryRepository
	modelRepo     ProductModelRepository
	imageRepo     ProductImageURLRepository
	meshyClient   ModelGenerationClient
	storage       StorageClient
	modelNotifier ModelNotifier
}

func NewProductService(
	r ProductRepository,
	h ViewingHistoryRepository,
	modelRepo ProductModelRepository,
	imageRepo ProductImageURLRepository,
	meshyClient ModelGenerationClient,
	storage StorageClient,
	notifier ModelNotifier,
) *ProductService {
	return &ProductService{
		repo:          r,
		historyRepo:   h,
		modelRepo:     modelRepo,
		imageRepo:     imageRepo,
		meshyClient:   meshyClient,
		storage:       storage,
		modelNotifier: notifier,
	}
}

func (s *ProductService) ListProducts(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error) {
	return s.repo.List(ctx, f)
}

func (s *ProductService) Create(ctx context.Context, sellerID string, input domain.ProductCreate) (domain.Product, error) {
	p, err := s.repo.Create(ctx, sellerID, input)
	if err != nil {
		return domain.Product{}, err
	}

	modelID, err := s.modelRepo.Create(ctx, p.ID)
	if err != nil {
		slog.Error("failed to create product model record", "productID", p.ID, "error", err)
		return p, nil
	}

	imageURLs, err := s.imageRepo.GetURLsByProductID(ctx, p.ID)
	if err != nil || len(imageURLs) == 0 {
		slog.Error("failed to get image URLs for model generation", "productID", p.ID, "error", err)
		_ = s.modelRepo.UpdateStatus(context.Background(), modelID, domain.ModelStatusFailed)
		return p, nil
	}

	go s.runModelGeneration(modelID, p.ID, sellerID, imageURLs)
	return p, nil
}

func (s *ProductService) Update(ctx context.Context, id string, sellerID string, input domain.ProductUpdate) error {
	return s.repo.Update(ctx, id, sellerID, input)
}

func (s *ProductService) Delete(ctx context.Context, id string, sellerID string) error {
	return s.repo.Delete(ctx, id, sellerID)
}

func (s *ProductService) ListBySeller(ctx context.Context, sellerID string, f domain.ListingsFilter) ([]domain.Product, int, error) {
	return s.repo.ListBySeller(ctx, sellerID, f)
}

func (s *ProductService) ListViewingHistory(ctx context.Context, userID string, f domain.ViewingHistoryFilter) ([]domain.ViewingHistory, int, error) {
	return s.historyRepo.ListByUserID(ctx, userID, f)
}

func (s *ProductService) GetByID(ctx context.Context, id string, uid *string) (domain.ProductDetail, error) {
	product, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		return domain.ProductDetail{}, err
	}
	if uid != nil {
		if err := s.historyRepo.Upsert(ctx, *uid, id); err != nil {
			slog.Warn("failed to upsert viewing history", "error", err)
		}
	}
	return product, nil
}

func (s *ProductService) runModelGeneration(modelID, productID, sellerUID string, imageURLs []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	jobID, err := s.meshyClient.CreateJob(ctx, imageURLs)
	if err != nil {
		slog.Error("meshy create job failed", "productID", productID, "error", err)
		s.markModelFailed(modelID, productID, sellerUID)
		return
	}

	if err := s.modelRepo.UpdateJobID(ctx, modelID, jobID); err != nil {
		slog.Error("failed to update model job_id", "modelID", modelID, "error", err)
		s.markModelFailed(modelID, productID, sellerUID)
		return
	}

	meshyGLBURL, ok := s.pollUntilDone(ctx, jobID)
	if !ok {
		slog.Error("model generation failed or timed out", "productID", productID, "jobID", jobID)
		s.markModelFailed(modelID, productID, sellerUID)
		return
	}

	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, meshyGLBURL, nil)
	if err != nil {
		slog.Error("failed to create GLB download request", "error", err)
		s.markModelFailed(modelID, productID, sellerUID)
		return
	}
	dlResp, err := http.DefaultClient.Do(dlReq)
	if err != nil {
		slog.Error("failed to download GLB from Meshy", "error", err)
		s.markModelFailed(modelID, productID, sellerUID)
		return
	}
	defer dlResp.Body.Close()

	gcsURL, err := s.storage.Upload(ctx, "product-models/"+modelID+".glb", dlResp.Body, "model/gltf-binary")
	if err != nil {
		slog.Error("failed to upload GLB to GCS", "productID", productID, "error", err)
		s.markModelFailed(modelID, productID, sellerUID)
		return
	}

	if err := s.modelRepo.UpdateDone(ctx, modelID, gcsURL); err != nil {
		slog.Error("failed to update model done", "modelID", modelID, "error", err)
		return
	}

	s.modelNotifier.NotifyModelGenerationComplete(sellerUID, ModelGenerationNotification{
		ProductID: productID,
		GlbURL:    gcsURL,
	})
}

func (s *ProductService) pollUntilDone(ctx context.Context, jobID string) (glbURL string, ok bool) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for i := 0; i < 60; i++ {
		select {
		case <-ctx.Done():
			return "", false
		case <-ticker.C:
		}
		status, url, err := s.meshyClient.GetJobStatus(ctx, jobID)
		if err != nil {
			slog.Warn("meshy poll error", "jobID", jobID, "error", err)
			continue
		}
		switch status {
		case "SUCCEEDED":
			return url, true
		case "FAILED", "EXPIRED":
			return "", false
		}
	}
	return "", false
}

func (s *ProductService) markModelFailed(modelID, productID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.modelRepo.UpdateStatus(ctx, modelID, domain.ModelStatusFailed); err != nil {
		slog.Error("failed to mark model as failed", "modelID", modelID, "error", err)
	}
	s.modelNotifier.NotifyModelGenerationFailed(userID, productID)
}
