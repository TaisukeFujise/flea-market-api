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
	userService := service.NewUserService(userRepo, fb, gcs)
	userHandler := handler.NewUserHandler(userService)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	imageRepo := repository.NewProductImageRepository(db)
	summaryRepo := repository.NewDamageDetectionSummaryRepository(db)
	imageService := service.NewImageService(gcs, imageRepo, summaryRepo)
	imageHandler := handler.NewImageHandler(imageService)

	productRepo := repository.NewProductRepository(db)
	viewingHistoryRepo := repository.NewViewingHistoryRepository(db)
	productService := service.NewProductService(productRepo, viewingHistoryRepo)
	productHandler := handler.NewProductHandler(productService)

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, productRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	commentRepo := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepo)
	commentHandler := handler.NewCommentHandler(commentService)

	likeRepo := repository.NewLikeRepository(db)
	likeService := service.NewLikeService(likeRepo)
	likeHandler := handler.NewLikeHandler(likeService)

	api := e.Group("/api")
	public := api.Group("")
	authed := api.Group("")
	authed.Use(authMW.AuthRequired)

	// users
	public.POST("/users/register", userHandler.Register, authMW.TokenOnly)
	authed.GET("/me", userHandler.Get)
	authed.PATCH("/me", userHandler.Update)
	authed.DELETE("/me", userHandler.Delete)
	authed.PUT("/me/avatar", userHandler.UploadAvatar)
	authed.GET("/me/likes", likeHandler.GetLikes)
	authed.GET("/me/viewing-history", productHandler.GetViewingHistory)

	// categories
	public.GET("/categories", categoryHandler.GetAll)

	// products
	public.GET("/products", productHandler.GetList)
	public.GET("/products/:id", productHandler.GetByID, authMW.TokenOptional)
	authed.POST("/products", productHandler.Create)
	authed.PATCH("/products/:id", productHandler.Update)
	authed.DELETE("/products/:id", productHandler.Delete)

	// images
	authed.POST("/images", imageHandler.Upload)

	// damages
	public.GET("/products/:id/damages", notImplemented)
	authed.PATCH("/damages/:id", notImplemented)

	// comments
	public.GET("/products/:id/comments", commentHandler.GetList)
	authed.POST("/products/:id/comments", commentHandler.Create)
	authed.DELETE("/comments/:id", commentHandler.Delete)

	// likes
	authed.POST("/products/:id/likes", likeHandler.Create)
	authed.DELETE("/products/:id/likes", likeHandler.Delete)

	// orders
	authed.POST("/products/:id/orders", orderHandler.Create)
	authed.GET("/orders", orderHandler.GetList)
	authed.GET("/orders/:id", orderHandler.GetByID)
	authed.PATCH("/orders/:id", orderHandler.UpdateStatus)
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
