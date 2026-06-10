package handler

import (
	"bytes"
	"context"
	"io"
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

const maxImageSize = 10 << 20 // 10MB

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
	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return apperror.ErrUnauthorized.New("unauthorized")
	}

	uploads := make([]service.ImageUpload, 0, len(imageAngles))
	for _, angle := range imageAngles {
		a := string(angle)
		fh, err := c.FormFile(a)
		if err != nil {
			return apperror.ErrValidation.Wrap(err, a+" image is required")
		}
		if fh.Size > maxImageSize {
			return apperror.ErrValidation.New(a + " image exceeds 10MB limit")
		}
		f, err := fh.Open()
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to open "+a+" image")
		}
		defer f.Close()

		buf := make([]byte, 512)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return apperror.ErrInternal.Wrap(err, "failed to read "+a+" image")
		}
		ct := http.DetectContentType(buf[:n])
		if ct != "image/jpeg" && ct != "image/png" {
			return apperror.ErrValidation.New(a + " image must be JPEG or PNG")
		}
		uploads = append(uploads, service.ImageUpload{
			Reader:      io.MultiReader(bytes.NewReader(buf[:n]), f),
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
