# Epic 30 — Deployment & Production Readiness

**Goal:** Containerize with Docker, create docker-compose for local dev, add health checks, and prepare for production deployment.
**Dependencies:** All previous epics
**Commit:** `feat: add Docker, docker-compose, and production readiness`

---

## Step 30.01 — Dockerfile

**Build:** Create `backend/Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/ForgeBase ./cmd/ForgeBase

# Runtime stage
FROM alpine:3.20
WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

# Copy binary
COPY --from=builder /app/ForgeBase /app/ForgeBase

# Non-root user
RUN adduser -D -u 1000 appuser
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

EXPOSE 8080

ENTRYPOINT ["/app/ForgeBase"]
```

---

## Step 30.02 — Docker Compose (Local Development)

**Build:** Create `backend/docker-compose.yml`:

```yaml
version: '3.8'

services:
  ForgeBase:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ForgeBase_HOST=0.0.0.0
      - ForgeBase_PORT=8080
      - DATABASE_URL=postgres://ForgeBase:ForgeBase@db:5432/ForgeBase?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - RABBITMQ_URL=amqp://guest:guest@mq:5672/
      - EMQX_BROKER_URL=tcp://emqx:1883
      - SMTP_HOST=mailhog
      - SMTP_PORT=1025
      - LOG_LEVEL=debug
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
      mq:
        condition: service_started
      emqx:
        condition: service_started

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=ForgeBase
      - POSTGRES_PASSWORD=ForgeBase
      - POSTGRES_DB=ForgeBase
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ForgeBase"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  mq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      - RABBITMQ_DEFAULT_USER=guest
      - RABBITMQ_DEFAULT_PASS=guest

  emqx:
    image: emqx/emqx:5.0
    ports:
      - "1883:1883"
      - "8083:8083"
      - "8084:8084"
      - "18083:18083"

  mailhog:
    image: mailhog/mailhog:latest
    ports:
      - "1025:1025"
      - "8025:8025"

volumes:
  postgres_data:
  redis_data:
```

---

## Step 30.03 — Production Docker Compose

**Build:** Create `backend/docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  ForgeBase:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ForgeBase_ENV=production
      - ForgeBase_LOG_LEVEL=warn
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=ForgeBase
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ForgeBase"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

---

## Step 30.04 — GitHub Actions CI/CD

**Build:** Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U test"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
      rabbitmq:
        image: rabbitmq:3-management-alpine
        ports:
          - 5672:5672

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Download dependencies
        run: go mod download

      - name: Lint
        run: make lint

      - name: Test
        run: make test
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379/0
          RABBITMQ_URL: amqp://guest:guest@localhost:5672/

      - name: Upload coverage
        uses: codecov/codecov-action@v4
```

---

## Step 30.05 — Health Check Endpoint

**Build:** Update health handler to include dependency status:

```go
// HealthCheck handles GET /health
func HealthCheck(c *gin.Context) {
    status := "ok"
    checks := gin.H{}

    // Check database
    if err := db.Ping(c.Request.Context()); err != nil {
        status = "degraded"
        checks["database"] = "error"
    } else {
        checks["database"] = "ok"
    }

    // Check Redis
    if err := cache.Ping(c.Request.Context()); err != nil {
        status = "degraded"
        checks["cache"] = "error"
    } else {
        checks["cache"] = "ok"
    }

    c.JSON(http.StatusOK, gin.H{
        "status":    status,
        "service":   "ForgeBase",
        "checks":    checks,
        "version":   version,
        "built_at":  builtAt,
    })
}
```

---

## Step 30.06 — Build Info

**Build:** Create `backend/internal/version/version.go`:

```go
package version

var (
    Version   = "dev"
    BuiltAt   = "unknown"
    GitCommit = "unknown"
)
```

Update build command in `Makefile`:
```makefile
build:
	go build -ldflags="-s -w -X github.com/muhammadyunus/ForgeBase/internal/version.Version=$(VERSION) -X github.com/muhammadyunus/ForgeBase/internal/version.BuiltAt=$(BUILT_AT)" -o bin/ForgeBase ./cmd/ForgeBase
```

**Test cases:**
- [ ] Unit: Docker build succeeds
- [ ] E2E: docker-compose up starts all services
- [ ] E2E: /health returns status with dependency checks
- [ ] CI: GitHub Actions pipeline passes all tests

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add Docker, docker-compose, CI/CD, and production readiness"
```
