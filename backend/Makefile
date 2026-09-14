.PHONY: run test tidy build compose-up compose-down migrate migrate-down run-local

BIN := bin/cga-api
DOCKER ?= $(shell command -v docker >/dev/null 2>&1 && echo yes)

ifneq (,$(wildcard .env))
include .env
export
endif

tidy:
	go mod tidy

build:
	go build -o $(BIN) ./cmd/cga-api

test:
	go test ./...

compose-up:
	docker compose up -d --wait postgres

compose-down:
	docker compose down

# 对 DATABASE_URL 执行 golang-migrate up（含 audit_logs）。make run 启动时也会自动 up。
migrate:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required for migrate" && exit 1)
	go run ./cmd/cga-migrate -dir up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required for migrate" && exit 1)
	go run ./cmd/cga-migrate -dir down

# docker compose 起 Postgres 后前台跑 API；无 Docker 时要求已配置 DATABASE_URL。
run:
ifeq ($(DOCKER),yes)
	docker compose up --build
else
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required when docker is unavailable" && exit 1)
	go run ./cmd/cga-api
endif

run-local:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" && exit 1)
	go run ./cmd/cga-api
