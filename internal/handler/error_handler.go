package handler

import (
	"errors"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/labstack/echo/v5"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Echo v5 の HTTPErrorHandler は (c, err) の順（v4 と逆）
func ErrorHandler(c *echo.Context, err error) {
	var appErr *apperror.AppError

	if !errors.As(err, &appErr) {
		appErr = apperror.ErrInternal.Wrap(err, "internal server error")
	}
	c.JSON(appErr.Code.HTTPStatus(), ErrorResponse{
		Error: ErrorBody{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		},
	})
}
