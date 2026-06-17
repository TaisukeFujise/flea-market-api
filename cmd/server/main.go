package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TaisukeFujise/flea-market-api/internal/infra/ai"
	"github.com/TaisukeFujise/flea-market-api/internal/infra/fbapp"
	"github.com/TaisukeFujise/flea-market-api/internal/infra/gcs"
	"github.com/TaisukeFujise/flea-market-api/internal/infra/postgres"
	"github.com/labstack/echo/v5"
)

func main() {
	db, err := postgres.NewDB()
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	authClient, err := fbapp.NewAuthClient(context.Background())
	if err != nil {
		slog.Error("failed to initialize firebase auth client", "error", err)
		os.Exit(1)
	}
	gcsClient, err := gcs.NewClient(context.Background())
	if err != nil {
		slog.Error("failed to initialize GCS client", "error", err)
		os.Exit(1)
	}
	defer gcsClient.Close()

	vertexAI, err := ai.NewVertexAIClient(context.Background())
	if err != nil {
		slog.Error("failed to initialize vertex ai client", "error", err)
		os.Exit(1)
	}
	defer vertexAI.Close()

	e := NewRouter(db, authClient, gcsClient, vertexAI)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sc := echo.StartConfig{
		Address:         ":8080",
		GracefulTimeout: 10 * time.Second,
	}
	if err := sc.Start(ctx, e); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
