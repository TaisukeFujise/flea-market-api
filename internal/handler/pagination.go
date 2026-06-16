package handler

import (
	"strconv"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/labstack/echo/v5"
)

func parsePagination(c *echo.Context, defaultLimit int) (limit, offset int, err error) {
	limit = defaultLimit
	if v := c.QueryParam("limit"); v != "" {
		n, parseErr := strconv.Atoi(v)
		if parseErr != nil || n <= 0 {
			return 0, 0, apperror.ErrValidation.New("invalid limit")
		}
		limit = min(n, 100)
	}
	if v := c.QueryParam("offset"); v != "" {
		n, parseErr := strconv.Atoi(v)
		if parseErr != nil || n < 0 {
			return 0, 0, apperror.ErrValidation.New("invalid offset")
		}
		offset = n
	}
	return limit, offset, nil
}

type productSummaryResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Price        int     `json:"price"`
	ThumbnailURL *string `json:"thumbnail_url"`
	Status       string  `json:"status"`
}
