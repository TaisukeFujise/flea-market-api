package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type damageRepository interface {
	ListByProductID(ctx context.Context, productID string) ([]domain.Damage, error)
}

type damageProductRepository interface {
	Exists(ctx context.Context, id string) (bool, error)
}

type DamageService struct {
	repo        damageRepository
	productRepo damageProductRepository
}

func NewDamageService(repo damageRepository, productRepo damageProductRepository) *DamageService {
	return &DamageService{repo: repo, productRepo: productRepo}
}

func (s *DamageService) ListByProductID(ctx context.Context, productID string) ([]domain.Damage, error) {
	exists, err := s.productRepo.Exists(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperror.ErrNotFound.New("product not found")
	}
	return s.repo.ListByProductID(ctx, productID)
}
