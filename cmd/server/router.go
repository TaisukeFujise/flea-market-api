package main

import (
	"database/sql"
	"os"

	"firebase.google.com/go/v4/auth"
	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/handler"
	mw "github.com/TaisukeFujise/flea-market-api/internal/middleware"
	"github.com/TaisukeFujise/flea-market-api/internal/repository"
	"github.com/TaisukeFujise/flea-market-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewRouter(db *sql.DB, fb *auth.Client) *echo.Echo {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.HTTPErrorHandler = handler.ErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	allowOrigins := []string{"*"}
	if origin := os.Getenv("FRONTEND_ORIGIN"); origin != "" {
		allowOrigins = []string{origin}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: allowOrigins}))

	authMW := mw.AuthMiddleware{Client: fb, DB: db}
	_ = authMW
	// - user
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, fb)
	userHandler := handler.NewUserHandler(userService)
	_ = userHandler
	return e
}

// func notImplemented(c *echo.Context) error {
// 	return c.NoContent(http.StatusNotImplemented)
// }

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return apperror.ErrValidation.Wrap(err, err.Error())
	}
	return nil
}
