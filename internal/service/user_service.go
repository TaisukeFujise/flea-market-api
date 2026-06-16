package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Register(ctx context.Context, user domain.User) error
	Update(ctx context.Context, id string, userUpdate domain.UserUpdate) error
	Get(ctx context.Context, id string) (domain.User, error)
	Delete(ctx context.Context, id string) error
	UpdateAvatar(ctx context.Context, id string, avatarURL string) error
}

type FirebaseClient interface {
	DeleteUser(ctx context.Context, uid string) error
}

type UserService struct {
	repo    UserRepository
	fb      FirebaseClient
	storage StorageClient
}

func NewUserService(r UserRepository, fb FirebaseClient, s StorageClient) *UserService {
	return &UserService{repo: r, fb: fb, storage: s}
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

func (s *UserService) UploadAvatar(ctx context.Context, id string, r io.Reader, contentType string) error {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	}
	name := fmt.Sprintf("avatars/%s%s", uuid.New().String(), ext)
	newURL, err := s.storage.Upload(ctx, name, r, contentType)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to upload avatar to GCS")
	}

	if err := s.repo.UpdateAvatar(ctx, id, newURL); err != nil {
		_ = s.storage.Delete(context.Background(), name)
		return err
	}

	if user.AvatarURL != nil {
		if oldName, ok := gcsObjectName(*user.AvatarURL); ok {
			_ = s.storage.Delete(context.Background(), oldName)
		}
	}

	return nil
}

// gcsObjectName extracts the object name from a GCS public URL.
// Returns ("", false) for non-GCS URLs (e.g. external OAuth profile images).
func gcsObjectName(url string) (string, bool) {
	const prefix = "https://storage.googleapis.com/"
	if !strings.HasPrefix(url, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(url, prefix)
	_, name, ok := strings.Cut(rest, "/")
	if !ok {
		return "", false
	}
	return name, true
}
