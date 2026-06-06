package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error)
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(r ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

func (s *ProductService) ListProducts(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error) {
	return s.repo.List(ctx, f)
}
