package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]domain.Category, error)
}

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(r CategoryRepository) *CategoryService {
	return &CategoryService{repo: r}
}

func (s *CategoryService) GetAll(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetAll(ctx)
}
