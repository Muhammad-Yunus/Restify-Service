.PHONY: build run test test-unit test-integration lint lint-fix fmt migrate-up migrate-down clean help

APP_NAME=ForgeBase
CMD=cmd/ForgeBase

build:
	go build -o bin/$(APP_NAME) ./$(CMD)

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

clean:
	rm -rf bin/ coverage.out

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:' Makefile | cut -d: -f1 | sort
