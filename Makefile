.PHONY: help up down dev dev-air run-api run-worker air-api air-worker build test lint migrate-up migrate-down migrate-create migrate-status db-create

# Load .env if present
ifneq (,$(wildcard .env))
    include .env
    export
endif

AIR ?= air

DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= cierge_user
DB_PASSWORD ?= cierge_pass
DB_NAME ?= subsync

POSTGRES_DSN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
GOOSE ?= go run github.com/pressly/goose/v3/cmd/goose@v3.24.1

help:
	@echo "SubSync development targets:"
	@echo "  make up            Start Redis (docker)"
	@echo "  make down          Stop containers"
	@echo "  make db-create     Create subsync database (if missing)"
	@echo "  make dev           up + db-create + migrate-up + run-api"
	@echo "  make dev-air       up + migrate-up + air-api (live reload)"
	@echo "  make run-api       Run API server"
	@echo "  make air-api       Run API with air live reload"
	@echo "  make run-worker    Run background worker"
	@echo "  make air-worker    Run worker with air live reload"
	@echo "  make build         Build api and worker binaries"
	@echo "  make test          Run unit tests"
	@echo "  make lint          Run golangci-lint"
	@echo "  make migrate-up    Apply database migrations"
	@echo "  make migrate-down  Roll back last migration"
	@echo "  make migrate-status Show migration status"

up:
	docker compose up -d

down:
	docker compose down

db-create:
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres \
		-tAc "SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)'" | grep -q 1 \
		|| PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres \
		-c "CREATE DATABASE $(DB_NAME);"
	@echo "Database $(DB_NAME) is ready."

dev: up db-create migrate-up run-api

dev-air: up migrate-up air-api

run-api:
	go run ./cmd/api

air-api:
	$(AIR) -c .air.toml

run-worker:
	go run ./cmd/worker

air-worker:
	$(AIR) -c .air.worker.toml

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run ./...

migrate-up:
	$(GOOSE) -dir migrations postgres "$(POSTGRES_DSN)" up

migrate-down:
	$(GOOSE) -dir migrations postgres "$(POSTGRES_DSN)" down

migrate-create:
	$(GOOSE) -dir migrations create $(name) sql

migrate-status:
	$(GOOSE) -dir migrations postgres "$(POSTGRES_DSN)" status
