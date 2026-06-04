include .env
export

.PHONY: up rebuild down down-v dev migrate migrate-down migrate-local migrate-local-down

up:
	docker compose up

rebuild:
	docker compose build --no-cache
	docker compose up

down:
	docker compose down

down-v:
	docker compose down -v

dev:
	go run ./cmd/server

migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down

migrate-local:
	migrate -path ./migrations -database "$(subst @db:,@localhost:,$(DATABASE_URL))" up

migrate-local-down:
	migrate -path ./migrations -database "$(subst @db:,@localhost:,$(DATABASE_URL))" down
