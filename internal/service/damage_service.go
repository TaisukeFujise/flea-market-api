package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type DamageService struct {
	repo DamageRepository
}

func NewDamageService(repo DamageRepository) *DamageService {
	return &DamageService{repo: repo}
}

func (s *DamageService) ListByProductID(ctx context.Context, productID string) ([]domain.Damage, error) {
	return s.repo.ListByProductID(ctx, productID)
}
