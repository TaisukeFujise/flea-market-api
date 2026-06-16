package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/labstack/echo/v5"
)

type UserService interface {
	Register(ctx context.Context, user domain.User) error
	Update(ctx context.Context, id string, userUpdate domain.UserUpdate) error
	Get(ctx context.Context, id string) (domain.User, error)
	Delete(ctx context.Context, id string) error
	UploadAvatar(ctx context.Context, id string, r io.Reader, contentType string) error
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(s UserService) *UserHandler {
	return &UserHandler{service: s}
}

func firebaseUID(c *echo.Context) (string, error) {
	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return "", apperror.ErrUnauthorized.New("unauthorized")
	}
	return uid, nil
}

type RegisterUserRequest struct {
	DisplayName string  `json:"display_name" validate:"required,max=255"`
	AvatarURL   *string `json:"avatar_url"   validate:"omitempty,http_url"`
}

func (u *UserHandler) Register(c *echo.Context) error {
	var req RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	user := domain.User{
		ID:          uid,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
	}
	ctx := c.Request().Context()
	if err := u.service.Register(ctx, user); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,max=255"`
}

func (u *UserHandler) Update(c *echo.Context) error {
	var req UpdateUserRequest
	id, err := firebaseUID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userUpdate := domain.UserUpdate{
		DisplayName: req.DisplayName,
	}
	if err := u.service.Update(ctx, id, userUpdate); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type GetUserResponse struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	RatingAvg   *float64  `json:"rating_avg"`
	RatingCount int       `json:"rating_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *UserHandler) Get(c *echo.Context) error {
	id, err := firebaseUID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	user, err := u.service.Get(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, GetUserResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		RatingAvg:   user.RatingAvg,
		RatingCount: user.RatingCount,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

func (u *UserHandler) UploadAvatar(c *echo.Context) error {
	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	fh, err := c.FormFile("avatar")
	if err != nil {
		return apperror.ErrValidation.Wrap(err, "avatar image is required")
	}
	if fh.Size > maxImageSize {
		return apperror.ErrValidation.New("avatar image exceeds 10MB limit")
	}

	f, err := fh.Open()
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to open avatar image")
	}
	defer f.Close()

	ct, r, err := sniffImage(f)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to read avatar image")
	}
	if ct != "image/jpeg" && ct != "image/png" {
		return apperror.ErrValidation.New("avatar image must be JPEG or PNG")
	}

	ctx := c.Request().Context()
	if err := u.service.UploadAvatar(ctx, uid, r, ct); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (u *UserHandler) Delete(c *echo.Context) error {
	id, err := firebaseUID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := u.service.Delete(ctx, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
