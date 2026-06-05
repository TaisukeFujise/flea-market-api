package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/service"
	"github.com/labstack/echo/v5"
)

type ImageService interface {
	UploadImages(ctx context.Context, userID string, uploads []service.ImageUpload) ([]string, error)
}

type ImageHandler struct {
	service ImageService
}

func NewImageHandler(s ImageService) *ImageHandler {
	return &ImageHandler{service: s}
}

const maxImageSize = 10 << 20 // 10MB

var imageAngles = []string{"front", "back", "right", "left", "top"}

type uploadImagesResponse struct {
	ImageIDs        []string `json:"image_ids"`
	DamageDetection string   `json:"damage_detection"`
}

func (h *ImageHandler) Upload(c *echo.Context) error {
	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return apperror.ErrUnauthorized.New("unauthorized")
	}

	uploads := make([]service.ImageUpload, 0, len(imageAngles))
	for _, angle := range imageAngles {
		fh, err := c.FormFile(angle)
		if err != nil {
			return apperror.ErrValidation.Wrap(err, angle+" image is required")
		}
		if fh.Size > maxImageSize {
			return apperror.ErrValidation.New(angle + " image exceeds 10MB limit")
		}
		ct := fh.Header.Get("Content-Type")
		if ct != "image/jpeg" && ct != "image/png" {
			return apperror.ErrValidation.New(angle + " image must be JPEG or PNG")
		}
		f, err := fh.Open()
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to open "+angle+" image")
		}
		defer f.Close()
		uploads = append(uploads, service.ImageUpload{
			Reader:      f,
			ContentType: ct,
			Angle:       angle,
		})
	}

	ctx := c.Request().Context()
	ids, err := h.service.UploadImages(ctx, uid, uploads)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, uploadImagesResponse{
		ImageIDs:        ids,
		DamageDetection: "processing",
	})
}
