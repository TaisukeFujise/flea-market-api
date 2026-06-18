package service

import (
	"context"
	"log/slog"

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

type ProductService struct {
	repo        ProductRepository
	historyRepo ViewingHistoryRepository
}

func NewProductService(r ProductRepository, h ViewingHistoryRepository) *ProductService {
	return &ProductService{repo: r, historyRepo: h}
}

func (s *ProductService) ListProducts(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error) {
	return s.repo.List(ctx, f)
}

func (s *ProductService) Create(ctx context.Context, sellerID string, input domain.ProductCreate) (domain.Product, error) {
	return s.repo.Create(ctx, sellerID, input)
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
