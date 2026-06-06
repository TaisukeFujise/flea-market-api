package main

import (
	"database/sql"
	"net/http"
	"os"

	"firebase.google.com/go/v4/auth"
	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/handler"
	gcsclient "github.com/TaisukeFujise/flea-market-api/internal/infra/gcs"
	mw "github.com/TaisukeFujise/flea-market-api/internal/middleware"
	"github.com/TaisukeFujise/flea-market-api/internal/repository"
	"github.com/TaisukeFujise/flea-market-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewRouter(db *sql.DB, fb *auth.Client, gcs *gcsclient.Client) *echo.Echo {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.HTTPErrorHandler = handler.ErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	if origin := os.Getenv("FRONTEND_ORIGIN"); origin != "" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{origin}}))
	}

	authMW := mw.AuthMiddleware{Client: fb, DB: db}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, fb)
	userHandler := handler.NewUserHandler(userService)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	imageRepo := repository.NewProductImageRepository(db)
	summaryRepo := repository.NewDamageDetectionSummaryRepository(db)
	imageService := service.NewImageService(gcs, imageRepo, summaryRepo)
	imageHandler := handler.NewImageHandler(imageService)

	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	api := e.Group("/api")
	public := api.Group("")
	authed := api.Group("")
	authed.Use(authMW.AuthRequired)

	// users
	public.POST("/users/register", userHandler.Register, authMW.TokenOnly)
	authed.GET("/me", userHandler.Get)
	authed.PATCH("/me", userHandler.Update)
	authed.DELETE("/me", userHandler.Delete)
	authed.GET("/me/likes", notImplemented)
	authed.GET("/me/viewing-history", notImplemented)

	// categories
	public.GET("/categories", categoryHandler.GetAll)

	// products
	public.GET("/products", productHandler.GetList)
	public.GET("/products/:id", notImplemented)
	authed.POST("/products", notImplemented)
	authed.PATCH("/products/:id", notImplemented)
	authed.DELETE("/products/:id", notImplemented)

	// images
	authed.POST("/images", imageHandler.Upload)

	// damages
	public.GET("/products/:id/damages", notImplemented)
	authed.PATCH("/damages/:id", notImplemented)

	// comments
	public.GET("/products/:id/comments", notImplemented)
	authed.POST("/products/:id/comments", notImplemented)
	authed.DELETE("/comments/:id", notImplemented)

	// likes
	authed.POST("/products/:id/likes", notImplemented)
	authed.DELETE("/products/:id/likes", notImplemented)

	// orders
	authed.POST("/products/:id/orders", notImplemented)
	authed.GET("/orders", notImplemented)
	authed.GET("/orders/:id", notImplemented)
	authed.PATCH("/orders/:id", notImplemented)
	authed.POST("/orders/:id/damage-reports", notImplemented)

	// message rooms
	authed.GET("/message-rooms/:id/messages", notImplemented)
	authed.POST("/message-rooms/:id/messages", notImplemented)

	// websocket
	e.GET("/ws", notImplemented)

	return e
}

func notImplemented(c *echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return apperror.ErrValidation.Wrap(err, err.Error())
	}
	return nil
}
