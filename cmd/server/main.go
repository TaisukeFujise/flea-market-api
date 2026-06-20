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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 前回のプロセス終了時に中断されたモデル生成ジョブを failed にリセットする。
	// ゴルーチンがシャットダウン前に完了できなかった場合の安全網。
	if _, err := db.ExecContext(ctx, `
		UPDATE product_models SET status = 'failed', updated_at = NOW()
		WHERE status IN ('pending', 'processing') AND deleted_at IS NULL
	`); err != nil {
		slog.Warn("failed to reset in-flight model jobs on startup", "error", err)
	}

	e := NewRouter(db, authClient, gcsClient, vertexAI)
	sc := echo.StartConfig{
		Address:         ":8080",
		GracefulTimeout: 10 * time.Second,
	}
	if err := sc.Start(ctx, e); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
