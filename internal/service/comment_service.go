package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type CommentRepository interface {
	ListByProductID(ctx context.Context, productID string, f domain.CommentFilter) ([]domain.Comment, int, error)
	Create(ctx context.Context, input domain.CommentCreate) (domain.Comment, error)
	Delete(ctx context.Context, id string, uid string) error
}

type CommentService struct {
	repo CommentRepository
}

func NewCommentService(r CommentRepository) *CommentService {
	return &CommentService{repo: r}
}

func (s *CommentService) ListByProductID(ctx context.Context, productID string, f domain.CommentFilter) ([]domain.Comment, int, error) {
	return s.repo.ListByProductID(ctx, productID, f)
}

func (s *CommentService) Create(ctx context.Context, input domain.CommentCreate) (domain.Comment, error) {
	return s.repo.Create(ctx, input)
}

func (s *CommentService) Delete(ctx context.Context, id string, uid string) error {
	return s.repo.Delete(ctx, id, uid)
}
