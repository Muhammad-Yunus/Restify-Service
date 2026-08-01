# Epic 01 — Project Foundation

**Goal:** Initialize Go module, create project skeleton, establish coding standards.
**Dependencies:** None (first epic)
**Commit:** `feat: initialize ForgeBase project structure`

---

## Step 01.01 — Go Module Initialization

**Build:**
```bash
cd backend
go mod init github.com/muhammadyunus/ForgeBase
go mod tidy
```

**Output files:**
- `backend/go.mod` — module declaration, Go 1.25+
- `backend/go.sum` — dependency checksums
- `backend/main.go` — minimal entry point (placeholder)

**Test cases:**
- [ ] Unit: `go mod tidy` succeeds without errors
- [ ] Unit: `go build ./...` compiles successfully
- [ ] Unit: `go vet ./...` reports no issues

---

## Step 01.02 — Project Directory Structure

**Build:** Create directory skeleton matching Standard Go Project Layout:

```
backend/
├── cmd/
│   └── ForgeBase/
│       └── main.go
├── internal/
│   ├── config/
│   ├── di/
│   ├── domain/
│   │   ├── entity/
│   │   ├── repository/
│   │   ├── service/
│   │   └── event/
│   ├── infrastructure/
│   │   ├── database/
│   │   ├── cache/
│   │   ├── queue/
│   │   ├── mqtt/
│   │   ├── messaging/
│   │   ├── websocket/
│   │   ├── logging/
│   │   ├── tracing/
│   │   ├── email/
│   │   └── auth/
│   ├── presentation/
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   ├── router/
│   │   │   └── dto/
│   │   └── websocket/
│   └── application/
│       ├── service/
│       ├── usecase/
│       ├── repository/
│       └── event/
├── pkg/
├── migrations/
├── docs/
├── test/
│   ├── integration/
│   └── unit/
├── scripts/
├── configs/
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── .gitignore
└── .golangci.yml
```

**Test cases:**
- [ ] Unit: All directories exist (shell check)
- [ ] Unit: No empty Go files with syntax errors
- [ ] Unit: `go build ./...` passes with new structure

---

## Step 01.03 — Entry Point (main.go)

**Build:** Create `backend/cmd/ForgeBase/main.go`:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/muhammadyunus/ForgeBase/internal/di"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    app, err := di Bootstrap(ctx)
    if err != nil {
        log.Fatalf("bootstrap failed: %v", err)
    }
    defer app.Close(ctx)

    if err := app.Run(ctx); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
```

**Stubs to create:**
- `internal/di/bootstrap.go` — returns a stub `App` interface
- `internal/di/app.go` — stub `App` with `Run()` and `Close()`

**Test cases:**
- [ ] Unit: `go build ./cmd/ForgeBase` compiles
- [ ] E2E: Application starts and exits cleanly on SIGTERM

---

## Step 01.04 — Makefile

**Build:** Create `backend/Makefile`:

```makefile
.PHONY: build run test lint clean migrate-up migrate-down help

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

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

clean:
	rm -rf bin/ coverage.out

help:
	@echo "Available targets:"
	@make -s help 2>/dev/null || true
```

**Test cases:**
- [ ] Unit: `make build` produces binary
- [ ] Unit: `make lint` runs without errors on clean code

---

## Step 01.05 — .gitignore

**Build:** Create `backend/.gitignore`:

```gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Vendor directory
vendor/

# Config files (local overrides)
configs/.local.*
.env.local

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Coverage
coverage.out
coverage.html

# Logs
*.log
logs/

# Docs
docs/swagger/*
!docs/swagger/.gitkeep
```

---

## Step 01.06 — .golangci.yml (Linting Configuration)

**Build:** Create `backend/.golangci.yml` with recommended rules for Clean Architecture:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - revive
    - gci
    - misspell
    - unparam
    - unconvert
    - predeclared
    - nosprintfhostport
    - bodyclose
    - contextcheck
    - dogsled
    - durationcheck
    - wrapcheck
    - wsl

linters-settings:
  gci:
    sections:
      - standard
      - default
      - prefix(github.com/muhammadyunus/ForgeBase)
  revive:
    rules:
      - name: exported
        arguments: ["disableStutteringCheck"]
      - name: package-comments
        disabled: true
      - name: blank-imports
      - name: context-as-argument
      - name: error-return
      - name: error-naming
      - name: if-return
  wrapcheck:
    ignoreSigs:
      - fmt.Errorf
      -pkg.New*
      -_
  errcheck:
    check-type-assertions: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - govet
    - path: mocks
      linters:
        - errcheck
```

---

## Step 01.07 — CLAUDE.md (Project Convention)

**Build:** Create root-level `CLAUDE.md`:

```markdown
# ForgeBase — Project Conventions

## Architecture
- Clean Architecture: domain → application → infrastructure → presentation
- Dependency Rule: inner layers know nothing of outer layers
- Manual DI via bootstrap container
- KISS: minimal abstractions, explicit code

## Code Style
- No comments on trivial code (let names speak)
- One comment per non-obvious WHY only
- Functional options for constructors with many params
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Context propagated to all IO operations

## Naming
- Entities: PascalCase, singular (`User`, `Workspace`)
- Repositories: `{Entity}Repository` interface, `{Entity}RepositoryImpl` struct
- Services: `{Entity}Service`
- DTOs: suffixed with `Request`/`Response`
- Handlers: suffixed with `Handler`
- Migrations: `NNNN_direction_description.sql`

## Testing
- Unit tests: `_test.go` co-located with source
- Integration tests: `test/integration/` with testcontainers
- Mocks: `test/mocks/` generated via go generate
- Coverage target: ≥ 80%

## Commits
- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `chore:`
- One epic per commit when possible
- Reference epic number in commit message when applicable
```

---

## Commit Instruction

```bash
cd backend
git add .
git commit -m "feat: initialize ForgeBase project foundation with Clean Architecture structure"
```
