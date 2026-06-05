package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type UserRepository interface {
	Register(ctx context.Context, user domain.User) error
	Update(ctx context.Context, id string, userUpdate domain.UserUpdate) error
	Get(ctx context.Context, id string) (domain.User, error)
	Delete(ctx context.Context, id string) error
}

type FirebaseClient interface {
	DeleteUser(ctx context.Context, uid string) error
}

type UserService struct {
	repo UserRepository
	fb   FirebaseClient
}

func NewUserService(r UserRepository, fb FirebaseClient) *UserService {
	return &UserService{repo: r, fb: fb}
}

func (s *UserService) Register(ctx context.Context, user domain.User) error {
	return s.repo.Register(ctx, user)
}

func (s *UserService) Update(ctx context.Context, id string, userUpdate domain.UserUpdate) error {
	return s.repo.Update(ctx, id, userUpdate)
}

func (s *UserService) Get(ctx context.Context, id string) (domain.User, error) {
	return s.repo.Get(ctx, id)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	// Firebase アカウントの削除を試みる。失敗しても DB 側の soft-delete は完了しているため
	// ミドルウェアがアクセスを拒否し続ける。孤立した Firebase アカウントは許容する。
	_ = s.fb.DeleteUser(ctx, id)
	return nil
}
