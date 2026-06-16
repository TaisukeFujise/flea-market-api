package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
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

var imageAngles = []domain.ImageAngle{
	domain.AngleFront,
	domain.AngleBack,
	domain.AngleRight,
	domain.AngleLeft,
	domain.AngleTop,
}

type uploadImagesResponse struct {
	ImageIDs        []string `json:"image_ids"`
	DamageDetection string   `json:"damage_detection"`
}

func (h *ImageHandler) Upload(c *echo.Context) error {
	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	uploads := make([]service.ImageUpload, 0, len(imageAngles))
	for _, angle := range imageAngles {
		name := string(angle)
		fh, err := c.FormFile(name)
		if err != nil {
			return apperror.ErrValidation.Wrap(err, name+" image is required")
		}
		if fh.Size > maxImageSize {
			return apperror.ErrValidation.New(name + " image exceeds 10MB limit")
		}
		f, err := fh.Open()
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to open "+name+" image")
		}
		defer f.Close()

		ct, r, err := sniffImage(f)
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to read "+name+" image")
		}
		if ct != "image/jpeg" && ct != "image/png" {
			return apperror.ErrValidation.New(name + " image must be JPEG or PNG")
		}
		uploads = append(uploads, service.ImageUpload{
			Reader:      r,
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
