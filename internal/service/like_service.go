package service

import (
	"context"
)

type LikeRepository interface {
	Create(ctx context.Context, productID, userID string) error
	Delete(ctx context.Context, productID, userID string) error
}

type LikeService struct {
	repo LikeRepository
}

func NewLikeService(r LikeRepository) *LikeService {
	return &LikeService{repo: r}
}

func (s *LikeService) Create(ctx context.Context, productID, userID string) error {
	return s.repo.Create(ctx, productID, userID)
}

func (s *LikeService) Delete(ctx context.Context, productID, userID string) error {
	return s.repo.Delete(ctx, productID, userID)
}
