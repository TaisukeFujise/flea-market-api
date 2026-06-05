# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Local development (requires DB running via docker compose)
make up              # Start app + PostgreSQL via Docker Compose
make dev             # Run server directly (requires DB already up)
make rebuild         # Cache-busting rebuild + start

# Database migrations
make migrate-local       # Apply migrations to local DB
make migrate-local-down  # Rollback local DB migrations

# Tests (CI uses go test ./...)
go test ./...

# Build check
go build ./...
```

Google ADC is required for Gemini / Vertex AI:
```bash
gcloud auth application-default login
```

Environment is configured via `.env` (copy from `.env.example`). The only required variable for local development is `DATABASE_URL`.

## Architecture

Clean Architecture layered as: `handler → service → repository`, wired in `cmd/server/router.go`.

```
cmd/server/
  main.go       — DB + Firebase init, graceful shutdown
  router.go     — DI wiring, Echo setup, CustomValidator

internal/
  handler/      — Echo handlers; own response structs; no business logic
  service/      — Business logic; depends on repository interfaces
  repository/   — SQL queries via database/sql + lib/pq
  domain/       — Domain types (currently minimal)
  middleware/   — Firebase token verification → sets "firebase_uid" in context
  apperror/     — Typed errors (ErrCode → HTTP status mapping)
  infra/
    postgres/   — sql.DB factory (reads DATABASE_URL)
    fbapp/      — Firebase Auth client factory (uses Google ADC)
```

**Framework**: Echo v5 (note: `HTTPErrorHandler` signature is `(c *echo.Context, err error)` — reversed from v4).

**Database**: PostgreSQL 17 + pgvector extension. Schema uses `users.id` as `VARCHAR(255)` (Firebase UID), all other PKs are UUID.

**Auth flow**: `Authorization: Bearer <Firebase ID Token>` → `middleware.AuthMiddleware.AuthRequired` verifies token, checks `users.deleted_at IS NULL`, sets `firebase_uid` in Echo context.

## Coding conventions

- Repository interfaces are defined in the **consumer package** (handler defines `UserService`, service defines `UserRepository`).
- Domain types carry no JSON tags — handlers map to their own request/response structs.
- `domain.XxxUpdate` uses all-pointer fields for PATCH operations; SQL uses `COALESCE($n, column)` to keep existing values when nil.
- Handler responses must use named structs — never `map[string]any`. Every JSON response shape is defined as a `xxxResponse` struct in the handler file.
- Domain ENUM values (condition, sort, status, etc.) are defined as typed string constants in `domain/` (e.g. `type ProductCondition string` + `const ConditionGood ProductCondition = "good"`). Use these constants in handlers and repositories instead of raw string literals.

## Error handling

`apperror.AppError` carries an `ErrCode` string and wraps the original error. Handlers never marshal `AppError` directly — `handler.ErrorHandler` (the Echo `HTTPErrorHandler`) converts it to `ErrorResponse{error: {code, message}}`. Use `ErrCode.New(msg)` or `ErrCode.Wrap(err, msg)` to create errors in service/repository layers.

## Validation

Request structs use `go-playground/validator/v10` tags. `CustomValidator` in `router.go` wraps validation failures as `apperror.ErrValidation`.

## Database schema highlights

Key ENUMs: `product_condition` (good/fair/poor), `product_status` (draft/on_sale/sold/deleted), `image_angle`, `damage_type`, `model_status`, `order_status`.

`feedback_embeddings.embedding` is `vector(1408)` — Vertex AI Multimodal Embedding dimension.

Soft deletes via `deleted_at TIMESTAMP` on users, products, product_images, product_models, damages, damage_reports, message_rooms, messages, comments.

## Planned endpoints (not yet implemented)

See `docs/api_spec.md` for the full planned API surface. Currently only the skeleton exists in `router.go`.
