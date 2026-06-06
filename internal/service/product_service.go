package service

import (
	"context"
	"log/slog"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error)
	GetByID(ctx context.Context, id string, uid *string) (domain.ProductDetail, error)
}

type ViewingHistoryRepository interface {
	Upsert(ctx context.Context, userID, productID string) error
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
