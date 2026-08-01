# Epic 05 — Repository Interfaces

**Goal:** Define all repository interfaces in the domain layer that infrastructure adapters will implement.
**Dependencies:** Epic 04 (Domain entities defined)
**Commit:** `feat: add repository interfaces in domain layer`

---

## Step 05.01 — User Repository Interface

**Build:** Create `backend/internal/domain/repository/user.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// UserRepository defines the data access contract for users.
type UserRepository interface {
    // Create inserts a new user and returns it with generated ID.
    Create(ctx context.Context, user *entity.User) error

    // FindByID returns a user by UUID, returns entity.ErrNotFound if not found.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

    // FindByEmail returns a user by email, returns entity.ErrNotFound if not found.
    FindByEmail(ctx context.Context, email string) (*entity.User, error)

    // Update partially updates a user (only non-nil fields).
    Update(ctx context.Context, user *entity.User) error

    // Delete marks a user as inactive (soft delete).
    Delete(ctx context.Context, id uuid.UUID) error

    // List returns paginated users.
    List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error)

    // CountActive returns the number of active users.
    CountActive(ctx context.Context) (int, error)
}
```

---

## Step 05.02 — Role Repository Interface

**Build:** Create `backend/internal/domain/repository/role.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// RoleRepository defines the data access contract for roles.
type RoleRepository interface {
    // Create inserts a new role.
    Create(ctx context.Context, role *entity.Role) error

    // FindByID returns a role by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)

    // FindByName returns a role by name.
    FindByName(ctx context.Context, name string) (*entity.Role, error)

    // List returns all roles.
    List(ctx context.Context) ([]*entity.Role, error)

    // Update partially updates a role.
    Update(ctx context.Context, role *entity.Role) error

    // Delete removes a role (only if not system role).
    Delete(ctx context.Context, id uuid.UUID) error

    // AssignUser assigns a role to a user.
    AssignUser(ctx context.Context, userID, roleID uuid.UUID) error

    // RemoveUser removes a role assignment from a user.
    RemoveUser(ctx context.Context, userID, roleID uuid.UUID) error

    // GetUserRoles returns all roles for a user.
    GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entity.Role, error)
}
```

---

## Step 05.03 — Workspace Repository Interface

**Build:** Create `backend/internal/domain/repository/workspace.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// WorkspaceRepository defines the data access contract for workspaces.
type WorkspaceRepository interface {
    // Create inserts a new workspace and returns it with generated ID.
    Create(ctx context.Context, ws *entity.Workspace) error

    // FindByID returns a workspace by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error)

    // FindBySlug returns a workspace by URL slug.
    FindBySlug(ctx context.Context, slug string) (*entity.Workspace, error)

    // Update partially updates a workspace.
    Update(ctx context.Context, ws *entity.Workspace) error

    // Delete removes a workspace.
    Delete(ctx context.Context, id uuid.UUID) error

    // List returns paginated workspaces owned by a user.
    List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error)

    // ListAll returns all workspaces (admin only).
    ListAll(ctx context.Context, page, pageSize int) ([]*entity.Workspace, int, error)

    // CountByOwner returns how many workspaces a user owns.
    CountByOwner(ctx context.Context, ownerID uuid.UUID) (int, error)
}
```

---

## Step 05.04 — Team Repository Interface

**Build:** Create `backend/internal/domain/repository/team.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// TeamRepository defines the data access contract for teams.
type TeamRepository interface {
    // Create inserts a new team.
    Create(ctx context.Context, team *entity.Team) error

    // FindByID returns a team by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Team, error)

    // ListByWorkspace returns all teams in a workspace.
    ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Team, error)

    // Update partially updates a team.
    Update(ctx context.Context, team *entity.Team) error

    // Delete removes a team.
    Delete(ctx context.Context, id uuid.UUID) error

    // AddMember adds a user to a team.
    AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error

    // RemoveMember removes a user from a team.
    RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error

    // GetMember returns a team member's info.
    GetMember(ctx context.Context, teamID, userID uuid.UUID) (*entity.TeamMember, error)

    // ListMembers returns all members of a team.
    ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error)

    // AssignWorkspace assigns a team to a workspace with a role.
    AssignWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, role entity.TeamWorkspaceRole) error

    // GetWorkspaceAccess returns all workspaces a team has access to.
    GetWorkspaceAccess(ctx context.Context, teamID uuid.UUID) ([]*entity.Workspace, error)
}
```

---

## Step 05.05 — Collection Repository Interface

**Build:** Create `backend/internal/domain/repository/collection.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// CollectionRepository defines the data access contract for collections.
type CollectionRepository interface {
    // Create inserts a new collection.
    Create(ctx context.Context, col *entity.Collection) error

    // FindByID returns a collection by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error)

    // FindBySlug returns a collection by slug within a workspace.
    FindBySlug(ctx context.Context, workspaceID, slug string) (*entity.Collection, error)

    // ListByWorkspace returns all collections in a workspace.
    ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error)

    // Update partially updates a collection.
    Update(ctx context.Context, col *entity.Collection) error

    // Delete removes a collection (and all its endpoints).
    Delete(ctx context.Context, id uuid.UUID) error

    // CountByWorkspace returns the number of collections in a workspace.
    CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error)
}
```

---

## Step 05.06 — Endpoint Repository Interface

**Build:** Create `backend/internal/domain/repository/endpoint.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// EndpointRepository defines the data access contract for endpoints.
type EndpointRepository interface {
    // Create inserts a new endpoint.
    Create(ctx context.Context, ep *entity.Endpoint) error

    // FindByID returns an endpoint by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error)

    // ListByCollection returns all endpoints in a collection.
    ListByCollection(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error)

    // ListByWorkspace returns all endpoints across all collections in a workspace.
    ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Endpoint, error)

    // Update partially updates an endpoint.
    Update(ctx context.Context, ep *entity.Endpoint) error

    // Delete removes an endpoint.
    Delete(ctx context.Context, id uuid.UUID) error

    // ToggleActive enables or disables an endpoint.
    ToggleActive(ctx context.Context, id uuid.UUID, active bool) error

    // FindByPath returns an endpoint matching a path and version.
    FindByPath(ctx context.Context, path, version string) (*entity.Endpoint, error)

    // CountByWorkspace returns the number of endpoints in a workspace.
    CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error)
}
```

---

## Step 05.07 — API Log Repository Interface

**Build:** Create `backend/internal/domain/repository/apilog.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// APILogRepository defines the data access contract for API logs.
type APILogRepository interface {
    // Create inserts a log entry.
    Create(ctx context.Context, log *entity.APILog) error

    // CreateBatch inserts multiple log entries.
    CreateBatch(ctx context.Context, logs []*entity.APILog) error

    // FindByID returns a log entry by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.APILog, error)

    // Search returns paginated logs matching filters.
    Search(ctx context.Context, workspaceID, endpointID uuid.UUID,
        level entity.LogLevel, method, path string,
        from, to time.Time, page, pageSize int) ([]*entity.APILog, int, error)

    // DeleteOlderThan removes logs older than the given date.
    DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)

    // CountByWorkspace returns log count for a workspace in a time range.
    CountByWorkspace(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (int64, error)
}
```

---

## Step 05.08 — Analytics & Alert Repository Interfaces

**Build:** Create `backend/internal/domain/repository/analytics.go`:

```go
package repository

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// AnalyticsRepository defines the data access contract for API analytics.
type AnalyticsRepository interface {
    // RecordMetric stores an aggregated metric.
    RecordMetric(ctx context.Context, metric *entity.AnalyticsMetric) error

    // RecordMetricsBatch stores multiple metrics.
    RecordMetricsBatch(ctx context.Context, metrics []*entity.AnalyticsMetric) error

    // GetMetrics returns metrics for a workspace in a time range.
    GetMetrics(ctx context.Context, workspaceID uuid.UUID, metricName string,
        from, to time.Time, interval time.Duration) ([]*entity.AnalyticsMetric, error)

    // GetEndpointMetrics returns metrics for a specific endpoint.
    GetEndpointMetrics(ctx context.Context, endpointID uuid.UUID,
        from, to time.Time) ([]*entity.AnalyticsMetric, error)

    // GetOverview returns summary metrics for dashboard.
    GetOverview(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (OverviewMetrics, error)

    // AggregateOldMetrics rolls up hourly metrics into daily for storage efficiency.
    AggregateOldMetrics(ctx context.Context, olderThan time.Time) error
}

// OverviewMetrics holds dashboard summary data.
type OverviewMetrics struct {
    TotalRequests   int64   `json:"total_requests"`
    AvgLatencyMs    float64 `json:"avg_latency_ms"`
    ErrorRate       float64 `json:"error_rate"`       // percentage
    TopEndpoints    []TopEndpoint `json:"top_endpoints"`
    RequestsByHour  []HourlyMetric `json:"requests_by_hour"`
}

type TopEndpoint struct {
    Path     string  `json:"path"`
    Requests int64   `json:"requests"`
    ErrorRate float64 `json:"error_rate"`
}

type HourlyMetric struct {
    Hour     time.Time `json:"hour"`
    Requests int64     `json:"requests"`
    Errors   int64     `json:"errors"`
}
```

**Build:** Create `backend/internal/domain/repository/alert.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// AlertRepository defines the data access contract for alert rules and events.
type AlertRepository interface {
    // Create inserts a new alert rule.
    Create(ctx context.Context, rule *entity.AlertRule) error

    // FindByID returns an alert rule by UUID.
    FindByID(ctx context.Context, id uuid.UUID) (*entity.AlertRule, error)

    // ListByWorkspace returns all alert rules for a workspace.
    ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.AlertRule, error)

    // Update partially updates an alert rule.
    Update(ctx context.Context, rule *entity.AlertRule) error

    // Delete removes an alert rule.
    Delete(ctx context.Context, id uuid.UUID) error

    // ToggleEnabled enables or disables a rule.
    ToggleEnabled(ctx context.Context, id uuid.UUID, enabled bool) error

    // CreateEvent records a fired alert event.
    CreateEvent(ctx context.Context, event *entity.AlertEvent) error

    // ListRecentEvents returns recent alert events for a workspace.
    ListRecentEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*entity.AlertEvent, error)

    // MarkNotified marks an alert event as notified.
    MarkNotified(ctx context.Context, id uuid.UUID) error
}
```

---

## Step 05.09 — Auth & Cache Repository Interfaces

**Build:** Create `backend/internal/domain/repository/auth.go`:

```go
package repository

import (
    "context"

    "github.com/google/uuid"
)

// TokenBlacklistRepository manages revoked JWT tokens.
type TokenBlacklistRepository interface {
    // Add adds a token to the blacklist.
    Add(ctx context.Context, tokenHash string, expiresAt time.Time) error

    // IsBlacklisted checks if a token is blacklisted.
    IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)

    // Cleanup removes expired blacklisted tokens.
    Cleanup(ctx context.Context) (int64, error)
}
```

**Build:** Create `backend/internal/domain/repository/cache.go`:

```go
package repository

import (
    "context"
    "time"
)

// Cache is the caching repository interface.
type Cache interface {
    // Get retrieves a value by key.
    Get(ctx context.Context, key string) (string, error)

    // Set stores a value with TTL.
    Set(ctx context.Context, key string, value string, ttl time.Duration) error

    // Delete removes a key.
    Delete(ctx context.Context, key string) error

    // Exists checks if a key exists.
    Exists(ctx context.Context, key string) (bool, error)

    // Close shuts down the cache connection.
    Close(ctx context.Context) error
}
```

---

## Step 05.10 — Infrastructure Interfaces (Logger, Queue, MQTT, Router)

**Build:** Create `backend/internal/domain/repository/logger.go`:

```go
package repository

import "log/slog"

// Logger is the logging repository interface.
type Logger interface {
    // With adds key-value pairs to the logger context.
    With(keyValues ...any) Logger

    // Debug logs a debug message.
    Debug(ctx context.Context, msg string, keyValues ...any)

    // Info logs an info message.
    Info(ctx context.Context, msg string, keyValues ...any)

    // Warn logs a warning message.
    Warn(ctx context.Context, msg string, keyValues ...any)

    // Error logs an error message.
    Error(ctx context.Context, msg string, keyValues ...any)

    // Logger returns the underlying *slog.Logger.
    Logger() *slog.Logger
}
```

**Build:** Create `backend/internal/domain/repository/queue.go`:

```go
package repository

import (
    "context"
)

// MessageQueue is the message queue repository interface.
type MessageQueue interface {
    // Publish publishes a message to a queue.
    Publish(ctx context.Context, queue string, message []byte) error

    // Consume starts consuming from a queue with the given handler.
    Consume(ctx context.Context, queue string, handler MessageHandler) error

    // DeclareQueue declares a queue with options.
    DeclareQueue(ctx context.Context, name string, opts QueueOptions) error

    // Close shuts down the queue connection.
    Close(ctx context.Context) error
}

// MessageHandler processes a received message.
type MessageHandler func(ctx context.Context, message []byte) error

// QueueOptions defines queue declaration parameters.
type QueueOptions struct {
    Durable   bool
    AutoDelete bool
    Arguments map[string]any
}
```

**Build:** Create `backend/internal/domain/repository/mqtt.go`:

```go
package repository

import "context"

// MQTTBroker is the MQTT broker repository interface.
type MQTTBroker interface {
    // Connect establishes connection to the broker.
    Connect(ctx context.Context) error

    // Publish publishes a message to a topic.
    Publish(ctx context.Context, topic string, payload []byte, qos byte, retained bool) error

    // Subscribe subscribes to a topic with a handler.
    Subscribe(ctx context.Context, topic string, qos byte, handler MQTTHandler) error

    // Unsubscribe unsubscribes from a topic.
    Unsubscribe(ctx context.Context, topic string) error

    // Close shuts down the MQTT connection.
    Close(ctx context.Context) error
}

// MQTTHandler processes a received MQTT message.
type MQTTHandler func(topic string, payload []byte)
```

**Build:** Create `backend/internal/domain/repository/router.go`:

```go
package repository

import "net/http"

// HTTPRouter defines the HTTP routing interface.
type HTTPRouter interface {
    // Group creates a route group with middleware.
    Group(basePath string, middleware ...Middleware) *RouterGroup

    // Handle registers a handler for a route.
    Handle(method, path string, handler http.HandlerFunc, middleware ...Middleware)

    // ServeHTTP implements http.Handler.
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// RouterGroup is a group of routes with shared middleware.
type RouterGroup struct {
    basePath string
    middleware []Middleware
}

// Middleware is a function that wraps an HTTP handler.
type Middleware func(http.HandlerFunc) http.HandlerFunc
```

---

## Step 05.11 — Compile-Time Interface Satisfaction Checks

**Build:** Create `backend/internal/domain/repository/checks.go`:

```go
package repository

// Compile-time checks to ensure infrastructure implementations satisfy interfaces.
// These will be satisfied after Epic 06.

// var _ DB = (*database.PostgresDB)(nil)       // Epic 03
// var _ Cache = (*cache.RedisCache)(nil)       // Epic 24
// var _ MessageQueue = (*queue.RabbitMQQueue)(nil)  // Epic 21
// var _ MQTTBroker = (*mqtt.EMQXBroker)(nil)   // Epic 22
// var _ Logger = (*logging.SLogger)(nil)       // Epic 18
// var _ HTTPRouter = (*router.GinRouter)(nil)  // Epic 13
```

---

## Step 05.12 — Domain Service Interfaces

**Build:** Create `backend/internal/domain/repository/service.go`:

```go
package repository

// This file documents service interfaces (defined in domain/service for implementation).
// Repository interfaces are in their own files above.
```

Create `backend/internal/domain/service/interface.go`:

```go
package service

import (
    "context"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
)

// AuthService handles authentication operations.
type AuthService interface {
    Register(ctx context.Context, email, password, fullName string) (*entity.User, error)
    Login(ctx context.Context, email, password string) (*AuthResult, error)
    RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
    Logout(ctx context.Context, token string) error
    GetCurrentUser(ctx context.Context, token string) (*entity.User, error)
    HasPermission(ctx context.Context, userID uuid.UUID, permission string) bool
}

// AuthResult holds the result of an authentication operation.
type AuthResult struct {
    User         *entity.User `json:"user"`
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresIn    int          `json:"expires_in"`
}

// UserService handles user management.
type UserService interface {
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
    Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.User, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error)
}

// WorkspaceService handles workspace management.
type WorkspaceService interface {
    Create(ctx context.Context, name, description string, ownerID uuid.UUID) (*entity.Workspace, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error)
    Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Workspace, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*entity.Workspace, int, error)
}

// TeamService handles team management.
type TeamService interface {
    Create(ctx context.Context, name string, workspaceID uuid.UUID) (*entity.Team, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error)
    AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
    RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
    ListMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.TeamMember, error)
    AssignToWorkspace(ctx context.Context, teamID, workspaceID uuid.UUID, wsRole entity.TeamWorkspaceRole) error
}

// CollectionService handles collection management.
type CollectionService interface {
    Create(ctx context.Context, name, description string, workspaceID uuid.UUID) (*entity.Collection, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error)
    Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Collection, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, workspaceID uuid.UUID) ([]*entity.Collection, error)
}

// EndpointService handles endpoint management.
type EndpointService interface {
    Create(ctx context.Context, collectionID uuid.UUID, params map[string]any) (*entity.Endpoint, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Endpoint, error)
    Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.Endpoint, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, collectionID uuid.UUID) ([]*entity.Endpoint, error)
    ToggleActive(ctx context.Context, id uuid.UUID, active bool) error
}

// APIIntrospector inspects database schemas and discovers objects.
type APIIntrospector interface {
    // DiscoverTables returns all user tables in a schema.
    DiscoverTables(ctx context.Context, schema string) ([]TableInfo, error)

    // DiscoverFunctions returns all user functions in a schema.
    DiscoverFunctions(ctx context.Context, schema string) ([]FunctionInfo, error)

    // DiscoverProcedures returns all user procedures in a schema.
    DiscoverProcedures(ctx context.Context, schema string) ([]ProcedureInfo, error)

    // GetTableSchema returns column information for a table.
    GetTableSchema(ctx context.Context, schema, table string) ([]ColumnSchema, error)

    // GetFunctionSignature returns parameter info for a function/procedure.
    GetFunctionSignature(ctx context.Context, schema, name string) ([]ParamSchema, error)
}

// TableInfo represents a discovered database table.
type TableInfo struct {
    Schema    string
    Name      string
    Columns   []ColumnSchema
    PrimaryKey []string
}

// ColumnSchema represents a column definition.
type ColumnSchema struct {
    Name      string
    Type      string
    IsNullable bool
    IsPrimary  bool
    Default   *string
}

// FunctionInfo represents a discovered database function.
type FunctionInfo struct {
    Schema   string
    Name     string
    Params   []ParamSchema
    ReturnType string
}

// ProcedureInfo represents a discovered database procedure.
type ProcedureInfo struct {
    Schema string
    Name   string
    Params []ParamSchema
}

// ParamSchema represents a function/procedure parameter.
type ParamSchema struct {
    Name    string
    Type    string
    Mode    string // IN, OUT, INOUT
}

// RESTGenerator generates REST endpoints from DB object bindings.
type RESTGenerator interface {
    // GenerateHandler creates an HTTP handler for an endpoint.
    GenerateHandler(ctx context.Context, endpoint *entity.Endpoint) (http.HandlerFunc, error)

    // ValidateBinding checks if the endpoint's DB binding is valid.
    ValidateBinding(ctx context.Context, endpoint *entity.Endpoint) error
}

// AnalyticsService manages API analytics data.
type AnalyticsService interface {
    RecordRequest(ctx context.Context, log *entity.APILog) error
    GetOverview(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (repository.OverviewMetrics, error)
    GetEndpointMetrics(ctx context.Context, endpointID uuid.UUID, from, to time.Time) ([]*entity.AnalyticsMetric, error)
}

// AlertService manages alert rules and notifications.
type AlertService interface {
    CreateRule(ctx context.Context, rule *entity.AlertRule) error
    GetRule(ctx context.Context, id uuid.UUID) (*entity.AlertRule, error)
    ListRules(ctx context.Context, workspaceID uuid.UUID) ([]*entity.AlertRule, error)
    UpdateRule(ctx context.Context, rule *entity.AlertRule) error
    DeleteRule(ctx context.Context, id uuid.UUID) error
    ToggleRule(ctx context.Context, id uuid.UUID, enabled bool) error
    FireAlert(ctx context.Context, event *entity.AlertEvent) error
}

// EmailService sends email notifications.
type EmailService interface {
    SendAlertEmail(ctx context.Context, recipient, subject, body string) error
    SendWelcomeEmail(ctx context.Context, recipient, name string) error
}

// SchedulerService manages scheduled background tasks.
type SchedulerService interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    // RegisterCron registers a job with a cron expression.
    RegisterCron(name string, cronExpr string, job func(context.Context) error) error
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add all repository and service interfaces in domain layer"
```
