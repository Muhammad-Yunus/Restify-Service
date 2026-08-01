# ForgeBase — REST API Builder & Runtime

> **ForgeBase** is a high-performance Go service that acts as a **REST API Builder and Runtime** (BAAS). It introspects PostgreSQL schemas, auto-generates CRUD endpoints, and provides analytics, observability, real-time subscriptions, and alerting — all out of the box.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/muhammadyunus/Restify-Service.svg)](https://pkg.go.dev/github.com/muhammadyunus/Restify-Service)
[![CI](https://github.com/muhammadyunus/Restify-Service/actions/workflows/ci.yml/badge.svg)](https://github.com/muhammadyunus/Restify-Service/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/muhammadyunus/Restify-Service/branch/main/graph/badge.svg)](https://codecov.io/gh/muhammadyunus/Restify-Service)
[![Go Report Card](https://goreportcard.com/badge/github.com/muhammadyunus/Restify-Service)](https://goreportcard.com/report/github.com/muhammadyunus/Restify-Service)
[![Release](https://img.shields.io/github/v/release/muhammadyunus/Restify-Service)](https://github.com/muhammadyunus/Restify-Service/releases)
[![Docker](https://img.shields.io/badge/Docker-ready-blue)](#-quick-start)

---

## Table of Contents

- [Architecture](#-architecture)
- [Core Flows](#-core-flows)
- [Project Structure](#-project-structure)
- [Quick Start](#-quick-start)
- [Build from Source](#-build-from-source)
- [Docker](#-docker)
- [Configuration](#-configuration)
- [API Documentation](#-api-documentation)
- [Example Usage](#-example-usage)
- [Testing](#-testing)
- [Migrations](#-migrations)
- [Security](#-security)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🏛️ Architecture

ForgeBase follows **Hexagonal (Ports & Adapters) Architecture** with strict layer separation:

```mermaid
graph TB
    subgraph "Presentation Layer"
        HTTP[HTTP Handler / Gin]
        WS[WebSocket Handler]
        SW[Swagger UI]
    end

    subgraph "Middleware Layer"
        MR[RequestID]
        RL[Rate Limiter]
        AL[Access Logger]
        Auth[JWT Auth]
    end

    subgraph "Application Layer"
        UC[Use Cases]
        Svc[Services]
        Job[Background Jobs]
    end

    subgraph "Domain Layer"
        Entity[Entities]
        Repo[Repository Ports]
        Event[Domain Events]
    end

    subgraph "Infrastructure Layer"
        DB[(PostgreSQL)]
        Cache[(Redis)]
        MQ[RabbitMQ]
        MQTT[(EMQX)]
        SMTP[SMTP]
        OTEL[OpenTelemetry]
    end

    HTTP --> MR --> RL --> Auth --> AL --> UC
    WS --> Auth --> SL
    SL[Service Layer] --> UC
    UC --> Entity
    UC --> Repo
    Repo -.-> DB
    Repo -.-> Cache
    Repo -.-> MQ
    MQ --> Job
    Job --> SMTP
    Job --> MQTT
    OTEL -.-> HTTP
    OTEL -.-> MQ
    OTEL -.-> DB
```

---

## 🔄 Core Flows

### Request Lifecycle

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Gin Router
    participant M as Middleware
    participant U as UseCase
    participant D as Domain Entity
    participant R as Repository
    participant DB as PostgreSQL

    C->>H: HTTP Request
    H->>M: Chain: RequestID → RateLimit → Auth → AccessLog
    M->>M: Validate JWT / Rate Check / Log ID
    H->>U: Route → Handler
    U->>D: Map DTO → Entity
    U->>R: Repository Call
    R->>DB: Query / Mutation
    DB-->>R: Row(s)
    R-->>U: Entity / Error
    U->>D: Map Entity → DTO
    U-->>H: JSON Response
    H-->>C: HTTP Response (200/201/4xx/5xx)
```

### BAAS Auto-Generation Flow

```mermaid
flowchart LR
    A[PostgreSQL<br/>Schema] --> B[Introspect<br/>Handler]
    B --> C[Generate<br/>OpenAPI Spec]
    C --> D[Register<br/>CRUD Routes]
    D --> E[RESTify<br/>Endpoints]
    E --> F[Auto CRUD<br/>APIs]

    style A fill:#e1f5fe
    style F fill:#c8e6c9
```

### Background Job Pipeline (RabbitMQ + gocron)

```mermaid
flowchart TD
    Trigger1[gocron Cron Schedule]
    Trigger2[Event Bus Publish]
    Trigger3[HTTP Queue Request]

    Trigger1 & Trigger2 & Trigger3 --> Queue[RabbitMQ Queue]
    Queue --> Worker[Worker Pool]
    Worker --> Process[Process / Transform]
    Process --> Notify[Push Notification<br/>via Email / WebSocket / MQTT]
    Process --> Persist[Save API Log / Analytics]

    style Queue fill:#fff3e0
    style Worker fill:#e8f5e9
    style Notify fill:#fce4ec
```

### WebSocket Real-Time Push

```mermaid
flowchart LR
    API[HTTP API Event] --> Event[Event Bus]
    Event --> Hub[WebSocket Hub]
    Hub --> Sub[Connected Subscribers]
    Sub --> Client1((Browser A))
    Sub --> Client2((Browser B))
    Sub --> Client3((Dashboard C))

    style Hub fill:#f3e5f5
    style API fill:#e3f2fd
```

---

## 📁 Project Structure

```
Restify-Service/
├── .github/
│   └── workflows/
│       └── ci.yml                  # CI: lint → test → coverage → build
├── backend/
│   ├── cmd/
│   │   └── ForgeBase/
│   │       └── main.go             # Application entry point
│   ├── configs/
│   │   └── .env.example             # Environment variable template
│   ├── docs/
│   │   ├── docs.go                  # Generated Swagger docs
│   │   ├── swagger.json             # OpenAPI 3.0 JSON spec
│   │   └── swagger/                 # Swagger UI assets
│   ├── internal/
│   │   ├── application/
│   │   │   ├── event/               # Domain event definitions
│   │   │   ├── repository/          # Application-level repository implementations
│   │   │   ├── service/             # Business services
│   │   │   └── usecase/             # Use case interactors
│   │   ├── config/                  # Configuration loading & validation
│   │   ├── di/                      # Dependency injection / IoC container
│   │   ├── domain/
│   │   │   ├── entity/              # Aggregate root entities
│   │   │   ├── repository/          # Repository interface ports
│   │   │   └── service/             # Domain service interfaces
│   │   ├── infrastructure/
│   │   │   ├── auth/                # JWT, password hashing, token blacklist
│   │   │   ├── baas/                # BaaS auto-generator (OpenAPI spec)
│   │   │   ├── cache/               # Redis cache adapter
│   │   │   ├── database/            # GORM + pgx PostgreSQL adapter
│   │   │   ├── email/               # SMTP email sender
│   │   │   ├── logging/             # Structured logging (tint)
│   │   │   ├── messaging/           # RabbitMQ amqp adapter
│   │   │   ├── mqtt/                # EMQX MQTT broker adapter
│   │   │   ├── presentation/
│   │   │   │   └── http/
│   │   │   │       ├── handler/     # Request handlers (auth, user, workspace, etc.)
│   │   │   │       ├── middleware/  # Auth, rate limit, request log, request ID
│   │   │   │       └── router/      # Gin router, dynamic route registration
│   │   │   ├── queue/               # RabbitMQ consumer/worker pool
│   │   │   ├── tracing/             # OpenTelemetry tracers/exporters
│   │   │   └── websocket/           # WebSocket hub & connections
│   │   └── version/                 # Build version info (ldflags injected)
│   ├── migrations/                  # SQL migrations (embedded at compile time)
│   ├── test/
│   │   ├── integration/             # Integration tests (DB + Redis required)
│   │   └── mocks/                   # Generated mock implementations
│   ├── Dockerfile                    # Multi-stage production Dockerfile
│   ├── docker-compose.yml            # Local dev stack (Postgres + Redis + RabbitMQ + EMQX)
│   ├── docker-compose.prod.yml       # Production deployment stack
│   ├── go.mod                        # Go module definition
│   ├── go.sum                        # Dependency checksums
│   ├── Makefile                      # Build/test/lint/swagger targets
│   └── README.md
├── docs/                              # Additional project documentation
├── epic/                              # Epic planning documents
├── agent/                             # AI agent configuration
├── CLAUDE.md                          # Claude agent instructions
├── AGENT.md                           # Agent documentation
├── LICENSE
└── README.md                          # ← This file
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.25+**
- **PostgreSQL 16+**
- **Redis 7+**
- **RabbitMQ 3.x** (for background jobs)
- **EMQX 5.x** (optional, for MQTT)

### 1. Clone & Configure

```bash
git clone https://github.com/muhammadyunus/Restify-Service.git
cd Restify-Service
cp backend/configs/.env.example backend/configs/.env
# Edit backend/configs/.env with your credentials
```

### 2. Start Infrastructure (Docker)

```bash
cd backend
docker compose up -d
```

This starts **PostgreSQL**, **Redis**, **RabbitMQ**, and **EMQX** on default ports.

### 3. Run Migrations

```bash
export DATABASE_URL="postgres://postgres:password@localhost:5432/forgebase?sslmode=disable"
make migration-up
# or
go run . migrate up
```

### 4. Build & Run

```bash
make build
./bin/ForgeBase
```

The server starts on `http://0.0.0.0:8080` by default.

### 5. Verify

```bash
curl http://localhost:8080/health
# → {"status":"ok","service":"ForgeBase","version":"1.14.0","built_at":"...","checks":{"cache":"ok","database":"ok"}}
```

---

## 🔨 Build from Source

```bash
cd backend

# Install dependencies
go mod download

# Run linter
make lint

# Format code
make fmt

# Run tests
make test                # Full suite with race detection
make test-unit           # Unit tests with coverage profile
make test-integration    # Integration tests (requires Postgres + Redis)

# Build production binary
make build               # Outputs to bin/ForgeBase with version ldflags

# Build with custom version
VERSION=2.0.0 BUILT_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)" make build

# Cleanup build artifacts
make clean
```

---

## 🐳 Docker

### Local Development

```bash
cd backend
docker compose up -d
make run
# Server running at http://localhost:8080
```

### Production Build

```bash
# Build image
docker build -t forgebase:latest -f backend/Dockerfile backend/

# Run with compose
docker compose -f backend/docker-compose.prod.yml up -d

# Or run directly
docker run -p 8080:8080 \
  -e FORGEBASE_DATABASE_URL=postgres://user:pass@db:5432/forgebase \
  -e FORGEBASE_REDIS_URL=redis://redis:6379/0 \
  forgebase:latest
```

### Docker Compose Services

| Service | Port | Purpose |
|---------|------|---------|
| `postgres` | `5432` | Primary relational database |
| `redis` | `6379` | Cache & session store |
| `rabbitmq` | `5672` / `15672` | Message queue (management UI on 15672) |
| `emqx` | `1883` / `18083` | MQTT broker (dashboard on 18083) |

---

## ⚙️ Configuration

All configuration is loaded from `configs/.env` (via [koanf](https://github.com/knadh/koanf)) and environment variables prefixed with `FORGEBASE_`.

### Required Variables

```env
# Server
FORGEBASE_SERVER_HOST=0.0.0.0
FORGEBASE_SERVER_PORT=8080
FORGEBASE_SERVER_ENV=development

# Database (PostgreSQL)
FORGEBASE_DATABASE_URL=postgres://user:pass@localhost:5432/forgebase?sslmode=disable

# Redis
FORGEBASE_REDIS_URL=redis://localhost:6379/0

# RabbitMQ
FORGEBASE_RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# JWT Auth
FORGEBASE_JWT_SECRET=your-super-secret-key-change-in-production
FORGEBASE_JWT_EXPIRATION=24h

# Rate Limiting
FORGEBASE_RATE_LIMIT_REQUESTS_PER_MINUTE=100

# Logging
FORGEBASE_LOGGING_LEVEL=info
FORGEBASE_LOGGING_FORMAT=json

# OpenTelemetry (optional)
FORGEBASE_OTEL_ENABLED=false
FORGEBASE_OTEL_ENDPOINT=http://localhost:4317

# EMQX MQTT (optional)
FORGEBASE_EMQX_URL=mqtts://localhost:8883

# SMTP (optional, for email alerts)
FORGEBASE_SMTP_HOST=smtp.gmail.com
FORGEBASE_SMTP_PORT=587
FORGEBASE_SMTP_USER=your@email.com
FORGEBASE_SMTP_PASS=your-password
```

### Configuration Priority

1. **Environment variables** (highest priority)
2. **`.env` file** at `configs/.env`
3. **Compiled defaults** (lowest priority)

---

## 📚 API Documentation

ForgeBase auto-generates an **OpenAPI 3.0** specification for all built-in endpoints and dynamically generated BAAS endpoints.

### Swagger UI

After starting the server, open:

```
http://localhost:8080/swagger/index.html
```

### Built-in API Endpoints

#### Auth (`/api/v1/auth`)
| Method | Path | Description | Access |
|--------|------|-------------|--------|
| `POST` | `/api/v1/auth/register` | Register new user | Public |
| `POST` | `/api/v1/auth/login` | Login & get JWT | Public |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | Auth |
| `POST` | `/api/v1/auth/logout` | Add token to blacklist | Auth |

#### Users (`/api/v1/users`)
| Method | Path | Description | Access |
|--------|------|-------------|--------|
| `GET` | `/api/v1/users` | List all users | Admin |
| `GET` | `/api/v1/users/:id` | Get user by ID | Admin |
| `PATCH` | `/api/v1/users/:id` | Update user | Admin |
| `DELETE` | `/api/v1/users/:id` | Delete user | Admin |

#### Workspaces (`/api/v1/workspaces`)
| Method | Path | Description | Access |
|--------|------|-------------|--------|
| `GET` | `/api/v1/workspaces` | List workspaces | Auth |
| `POST` | `/api/v1/workspaces` | Create workspace | Auth |
| `GET` | `/api/v1/workspaces/:id` | Get workspace | Auth |
| `PATCH` | `/api/v1/workspaces/:id` | Update workspace | Auth |
| `DELETE` | `/api/v1/workspaces/:id` | Delete workspace | Auth |

#### Teams (`/api/v1/teams/:id/members`)
| Method | Path | Description | Access |
|--------|------|-------------|--------|
| `POST` | `/api/v1/teams/:id/members` | Add team member | Admin |
| `GET` | `/api/v1/teams/:id/members` | List team members | Auth |
| `DELETE` | `/api/v1/teams/:id/members/:user_id` | Remove member | Admin |

#### Collections & Endpoints
| Method | Path | Description | Access |
|--------|------|-------------|--------|
| `GET`/`POST` | `/api/v1/collections` | List / Create collections | Auth |
| `PATCH`/`DELETE` | `/api/v1/collections/:id` | Update / Delete collection | Auth |
| `GET`/`POST` | `/api/v1/endpoints` | List / Create endpoints | Auth |
| `PATCH`/`DELETE` | `/api/v1/endpoints/:id` | Update / Delete endpoint | Auth |
| `POST` | `/api/v1/endpoints/:id/toggle` | Enable / Disable endpoint | Auth |
| `GET` | `/api/v1/introspect/schemas/:schema/tables` | Discover DB tables | Auth |
| `GET` | `/api/v1/introspect/schemas/:schema/tables/:table` | Get table schema | Auth |
| `GET` | `/api/v1/introspect/schemas/:schema/functions` | Discover functions | Auth |
| `GET` | `/api/v1/introspect/schemas/:schema/functions/:name` | Get function signature | Auth |

#### Analytics & Alerts
| Method | Path | Description | Access |
|--------|------|-------------|--------|
| `GET` | `/api/v1/analytics/overview/:workspace_id` | Analytics overview | Auth |
| `GET` | `/api/v1/analytics/endpoints/:endpoint_id/metrics` | Endpoint metrics | Auth |
| `GET` | `/api/v1/analytics/logs` | Search API logs | Admin |
| `GET`/`POST` | `/api/v1/alerts/:workspace_id` | List / Create alerts | Auth |
| `PUT` | `/api/v1/alerts/:workspace_id/:id/toggle` | Toggle alert | Auth |
| `PATCH` | `/api/v1/alerts/:workspace_id/:id` | Update alert | Auth |
| `DELETE` | `/api/v1/alerts/:workspace_id/:id` | Delete alert | Auth |
| `GET` | `/api/v1/alerts/:workspace_id/events` | Alert events history | Auth |

#### Dynamic BAAS Endpoints
| Pattern | Description |
|---------|-------------|
| `GET`/`POST` | `/api/:version/*path` | Auto-generated CRUD from DB schema |
| `PATCH`/`DELETE` | `/api/:version/*path` | Update / delete records |

#### WebSocket
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/ws` | Real-time event feed |

#### Health
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Service health & dependency checks |

---

## 💡 Example Usage

### Register & Login

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Secur3P@ss!"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Secur3P@ss!"}'
# → {"access_token":"eyJ...","refresh_token":"eyJ...","expires_at":"2026-..."}
```

### Create Workspace & Collection

```bash
TOKEN="eyJ..."

# Create workspace
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My API"!"}

# Discover tables
curl http://localhost:8080/api/v1/introspect/schemas/public/tables \
  -H "Authorization: Bearer $TOKEN"
# → [{"name":"users","row_count":10}, ...]

# Get table schema
curl http://localhost:8080/api/v1/introspect/schemas/public/tables/users \
  -H "Authorization: Bearer $TOKEN"
```

### Auto-Generate Endpoints

```bash
# Create endpoint from a table
curl -X POST http://localhost:8080/api/v1/endpoints \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "uuid-here",
    "collection_id": "uuid-here",
    "table_name": "users",
    "path": "/users",
    "methods": ["GET","POST","PATCH","DELETE"]
  }'

# Now the endpoint is live!
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

### Real-Time Events (WebSocket)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.addEventListener('open', () => {
  ws.send(JSON.stringify({ type: 'subscribe', topic: 'api.events' }));
});
ws.addEventListener('message', (e) => {
  console.log('Event:', JSON.parse(e.data));
});
```

### Health Check

```bash
curl http://localhost:8080/health
# → {
#   "status": "ok",
#   "service": "ForgeBase",
#   "version": "1.0.0",
#   "built_at": "2026-08-01T10:00:00Z",
#   "git_commit": "a1b2c3d",
#   "checks": {
#     "database": "ok",
#     "cache": "ok"
#   }
# }
```

### Rate Limiting Example

```bash
# Default: 100 requests/minute per IP
# Exceed → 429 Too Many Requests
for i in {1..105}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/v1/health
done
# Last 5 requests return 429
```

---

## 🧪 Testing

```bash
# All tests with race detector
make test

# Unit tests with coverage
make test-unit
# Coverage report: coverage.out → open with: go tool cover -html=coverage.out

# Integration tests
make test-integration

# Generate test coverage badge (for CI)
go test ./... -coverprofile=coverage.out
```

### Test Matrix

| Layer | Coverage | Commands |
|-------|----------|----------|
| Auth | ✅ Unit + Integration | `auth_test.go`, `jwt_test.go` |
| Cache | ✅ Unit + Integration | `cache_test.go` |
| Database | ✅ Unit + Integration | `database_test.go` |
| Entity | ✅ Unit (value objects) | `*_test.go` in `entity/` |
| Handler | ✅ Unit | `*_test.go` in `handler/` |
| Middleware | ✅ Unit | `auth_test.go`, `rate_limit_test.go` |
| Queue | ✅ Unit + Integration | `queue_test.go` |
| MQTT | ✅ Unit + Integration | `mqtt_test.go` |
| DI | ✅ Integration | `bootstrap_test.go` |

---

## 🗄️ Migrations

Migrations are embedded into the binary at compile time using Go `//go:embed`.

```bash
# Run all pending migrations
make migration-up

# Roll back last migration
make migration-down

# Custom DATABASE_URL
DATABASE_URL=postgres://user:pass@host:5432/db make migration-up
```

### Available Migrations

| # | File | Description |
|---|------|-------------|
| 001 | `create_users_table` | Users table (id, email, password_hash, name, is_admin, created_at, updated_at) |
| 002 | `create_roles_table` | Roles table (role_id, role_name, description) |
| 003 | `create_user_roles_table` | Junction table: user_id ↔ role_id |
| 004 | `create_workspaces_table` | Workspaces table (id, name, description, owner_id, created_at, updated_at) |
| 005 | `create_teams_table` | Teams table (id, workspace_id, name, created_at, updated_at) |
| 006 | `create_team_memberships_table` | Junction: team_id ↔ user_id |
| 007 | `create_workspace_teams_table` | Junction: workspace_id ↔ team_id |
| 008 | `create_collections_table` | Collections table (id, workspace_id, name, description, created_at, updated_at) |
| 009 | `create_endpoints_table` | Endpoints table (id, collection_id, workspace_id, name, path, methods, is_active, created_at, updated_at) |
| 010 | `create_api_logs_table` | API logs (request_id, endpoint_id, method, path, status, latency_ms, ip, user_agent, request_body, response_body, created_at) |
| 011 | `create_api_analytics_table` | Analytics aggregation (endpoint_id, period_start, request_count, error_count, avg_latency, p99_latency) |
| 012 | `create_alerts_table` | Alerts (workspace_id, endpoint_id, condition, threshold, notification_channels, is_active, created_at, updated_at) |
| 013 | `create_jwt_blacklist_table` | JWT blacklist (jti, expires_at) |

---

## 🔒 Security

| Feature | Implementation |
|---------|---------------|
| **Auth** | JWT (HS256) with configurable expiration |
| **Password** | bcrypt with cost factor 12 |
| **Token Revocation** | JWT blacklist stored in Redis (atomic set expiry) |
| **Rate Limiting** | 60s sliding window per IP (Redis-backed) |
| **CORS** | Configurable origins, methods, headers |
| **Request Logging** | UUID-traced, structured JSON logs with PII redaction |
| **Input Validation** | `go-playground/validator` on all DTO fields |
| **SQL Injection** | GORM parameterised queries (no raw SQL for inserts/updates) |

---

## 📈 Observability

### OpenTelemetry Tracing

Traces are emitted to an OTLP endpoint (gRPC) when `FORGEBASE_OTEL_ENABLED=true`.

```
Span flow:  HTTP Request → [Auth] → [Handler] → [Repository] → DB
              └─────────────── trace ID ───────────────────────┘
```

### Health & Dependency Checks

`GET /health` responds with:
- `status: "ok"` — all dependencies healthy
- `status: "degraded"` — one or more dependencies are unhealthy
- `checks` object with per-service status (`database`, `cache`)

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Commit changes (`git commit -am 'feat: add your feature'`)
4. Push to the branch (`git push origin feat/your-feature`)
5. Open a Pull Request

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(auth): add refresh token endpoint
fix(router): fix rate limiter key collision
chore: bump go.mod to 1.25
docs(readme): add architecture diagram
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgements

- [Gin](https://github.com/gin-gonic/gin) — HTTP web framework
- [GORM](https://gorm.io) — ORM for PostgreSQL
- [go-redis](https://github.com/redis/go-redis) — Redis client
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) — RabbitMQ client
- [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) — MQTT client
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go) — Observability SDK
- [koanf](https://github.com/knadh/koanf) — Configuration management
- [swaggo](https://github.com/swaggo/swag) — Swagger/OpenAPI generation
