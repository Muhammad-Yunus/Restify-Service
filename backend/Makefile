.PHONY: build run test test-unit test-integration lint lint-fix fmt migrate-up migrate-down clean swagger swagger-init swagger-fmt help

APP_NAME=ForgeBase
CMD=cmd/ForgeBase
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILT_AT ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags="-s -w -X github.com/muhammadyunus/Restify-Service/internal/version.Version=$(VERSION) -X github.com/muhammadyunus/Restify-Service/internal/version.BuiltAt=$(BUILT_AT) -X github.com/muhammadyunus/Restify-Service/internal/version.GitCommit=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")" -o bin/$(APP_NAME) ./$(CMD)

run: build
	./bin/$(APP_NAME)

test:
	go test ./... -race -count=1

test-unit:
	go test ./test/unit/... -coverprofile=coverage.out

test-integration:
	go test ./test/integration/... -race

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	golangci-lint fmt ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

swagger: swagger-init
	@echo "Swagger docs generated in docs/"

swagger-init:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || (echo "Installing swag CLI..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init --parseDepth 4 --parseInternal --output ./docs
	@echo "Post-processing generated docs.go to remove unused fields..."
	@if [ -f docs/docs.go ]; then \
		sed -i.bak '/LeftDelim:/d; /RightDelim:/d' docs/docs.go && rm docs/docs.go.bak; \
		echo "docs.go cleaned"; \
	fi

swagger-fmt:
	@echo "Formatting Swagger annotations..."
	swag fmt

clean:
	rm -rf bin/ coverage.out

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:' Makefile | cut -d: -f1 | sort
