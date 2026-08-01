# Epic 04 — Domain Entities

**Goal:** Define all core domain entities as Go structs with proper tags, validation rules, and domain methods.
**Dependencies:** Epic 03 (Database layer ready for entity mapping)
**Commit:** `feat: add domain entities for users, roles, workspaces, teams, collections, endpoints`

---

## Step 04.01 — User Entity

**Build:** Create `backend/internal/domain/entity/user.go`:

```go
package entity

import (
    "errors"
    "regexp"
    "strings"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/google/uuid"
)

// ErrNotFound is returned when a record is not found in the database.
var ErrNotFound = errors.New("record not found")

// User represents an application user.
type User struct {
    ID           uuid.UUID    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Email        string       `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
    PasswordHash string       `json:"-" gorm:"type:varchar(255);not null"`
    FullName     *string      `json:"full_name,omitempty" gorm:"type:varchar(255)"`
    IsActive     bool         `json:"is_active" gorm:"default:true"`
    CreatedAt    time.Time    `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt    time.Time    `json:"updated_at" gorm:"autoUpdateTime;not null"`

    // Relations
    Roles []*Role `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

// SetPassword hashes the plaintext password.
func (u *User) SetPassword(password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return err
    }
    u.PasswordHash = string(hash)
    return nil
}

// CheckPassword verifies the plaintext against the hash.
func (u *User) CheckPassword(password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// HasRole returns true if the user has the given role name.
func (u *User) HasRole(roleName string) bool {
    for _, r := range u.Roles {
        if r.Name == roleName {
            return true
        }
    }
    return false
}

// Validate checks user field constraints.
func (u *User) Validate() error {
    return validator.New().Struct(u)
}

// EmailValid returns whether the email format is valid.
func EmailValid(email string) bool {
    return validator.New().Var(email, "email") == nil
}
```

**Test cases:**
- [ ] Unit: `SetPassword()` produces non-empty hash
- [ ] Unit: `CheckPassword()` returns true for correct password
- [ ] Unit: `CheckPassword()` returns false for wrong password
- [ ] Unit: `HasRole()` returns true when role exists
- [ ] Unit: `HasRole()` returns false when role missing
- [ ] Unit: `Validate()` rejects empty email
- [ ] Unit: `Validate()` rejects invalid email format
- [ ] Unit: `EmailValid()` validates email format

---

## Step 04.02 — Role Entity

**Build:** Create `backend/internal/domain/entity/role.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// Role represents an authorization role.
type Role struct {
    ID          uuid.UUID   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name        string      `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
    Description *string     `json:"description,omitempty" gorm:"type:text"`
    IsSystem    bool        `json:"is_system" gorm:"default:false"` // system roles cannot be deleted
    CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime;not null"`

    // Relations
    Users []*User `json:"users,omitempty" gorm:"many2many:user_roles;"`
}

// System roles
const (
    RoleAdministrator = "administrator"
    RoleDeveloper     = "developer"
    RoleViewer        = "viewer"
    RoleTeamManager   = "team_manager"
)

// Validate checks role field constraints.
func (r *Role) Validate() error {
    return validator.New().Struct(r)
}
```

---

## Step 04.03 — Workspace Entity

**Build:** Create `backend/internal/domain/entity/workspace.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// Workspace represents a top-level isolation container.
type Workspace struct {
    ID          uuid.UUID   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name        string      `json:"name" gorm:"type:varchar(255);not null"`
    Description *string     `json:"description,omitempty" gorm:"type:text"`
    Slug        string      `json:"slug" gorm:"type:varchar(255);uniqueIndex;not null"`
    OwnerID     uuid.UUID   `json:"owner_id" gorm:"type:uuid;not null;index"`
    Owner       *User       `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
    IsPublic    bool        `json:"is_public" gorm:"default:false"`
    CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime;not null"`

    // Relations
    Collections []*Collection `json:"collections,omitempty" gorm:"foreignKey:WorkspaceID"`
    Teams       []*Team       `json:"teams,omitempty" gorm:"many2many:workspace_teams;"`
}

// GenerateSlug creates a URL-friendly slug from the name.
func (w *Workspace) GenerateSlug() {
    w.Slug = strings.ToLower(strings.ReplaceAll(w.Name, " ", "-"))
    re := regexp.MustCompile(`[^a-z0-9-]`)
    w.Slug = re.ReplaceAllString(w.Slug, "")
    re2 := regexp.MustCompile(`-+`)
    w.Slug = re2.ReplaceAllString(w.Slug, "-")
    w.Slug = strings.Trim(w.Slug, "-")
}

// Validate checks workspace field constraints.
func (w *Workspace) Validate() error {
    return validator.New().Struct(w)
}
```

---

## Step 04.04 — Team Entity

**Build:** Create `backend/internal/domain/entity/team.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// TeamWorkspaceRole defines the permission level a team has in a workspace.
type TeamWorkspaceRole string

const (
    TeamRoleViewer   TeamWorkspaceRole = "viewer"
    TeamRoleReadWrite TeamWorkspaceRole = "read_write"
    TeamRoleAdmin    TeamWorkspaceRole = "admin"
)

// Team represents a group of users.
type Team struct {
    ID        uuid.UUID       `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name      string          `json:"name" gorm:"type:varchar(255);not null"`
    WorkspaceID uuid.UUID     `json:"workspace_id" gorm:"type:uuid;not null;index"`
    Workspace *Workspace      `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
    CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`

    // Relations
    Members []*TeamMember `json:"members,omitempty" gorm:"foreignKey:TeamID"`
}

// Validate checks team field constraints.
func (t *Team) Validate() error {
    return validator.New().Struct(t)
}

// TeamMember represents a user's membership in a team.
type TeamMember struct {
    ID       uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    TeamID   uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
    UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
    Team     *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`
    User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Role     string    `json:"role" gorm:"type:varchar(50);default:'member'"`
    JoinedAt time.Time `json:"joined_at" gorm:"autoCreateTime;not null"`

    // Unique constraint on (team_id, user_id)
    // NOTE: GORM composite unique index is handled via migration SQL
}
```

---

## Step 04.05 — Collection Entity

**Build:** Create `backend/internal/domain/entity/collection.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// Collection groups related endpoints under a common namespace.
type Collection struct {
    ID          uuid.UUID   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name        string      `json:"name" gorm:"type:varchar(255);not null"`
    Description *string     `json:"description,omitempty" gorm:"type:text"`
    Slug        string      `json:"slug" gorm:"type:varchar(255);not null;index"`
    WorkspaceID uuid.UUID   `json:"workspace_id" gorm:"type:uuid;not null;index"`
    Workspace   *Workspace  `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
    CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime;not null"`

    // Relations
    Endpoints []*Endpoint `json:"endpoints,omitempty" gorm:"foreignKey:CollectionID"`
}

// GenerateSlug creates a URL-friendly slug.
func (c *Collection) GenerateSlug() {
    c.Slug = strings.ToLower(strings.ReplaceAll(c.Name, " ", "-"))
    re := regexp.MustCompile(`[^a-z0-9-]`)
    c.Slug = re.ReplaceAllString(c.Slug, "")
    re2 := regexp.MustCompile(`-+`)
    c.Slug = re2.ReplaceAllString(c.Slug, "-")
    c.Slug = strings.Trim(c.Slug, "-")
}

// Validate checks collection field constraints.
func (c *Collection) Validate() error {
    return validator.New().Struct(c)
}
```

---

## Step 04.06 — Endpoint Entity

**Build:** Create `backend/internal/domain/entity/endpoint.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// EndpointType defines the type of database object an endpoint targets.
type EndpointType string

const (
    EndpointTypeTable     EndpointType = "table"
    EndpointTypeFunction  EndpointType = "function"
    EndpointTypeProcedure EndpointType = "procedure"
)

// OperationType defines allowed HTTP operations.
type OperationType string

const (
    OpSelect   OperationType = "select"
    OpInsert   OperationType = "insert"
    OpUpdate   OperationType = "update"
    OpDelete   OperationType = "delete"
    OpCustom   OperationType = "custom"
)

// SecurityPolicy defines access control for an endpoint.
type SecurityPolicy struct {
    AuthRequired bool     `json:"auth_required"`
    AllowedRoles []string `json:"allowed_roles,omitempty"`
    RateLimit    *int     `json:"rate_limit,omitempty"` // requests per minute
}

// Endpoint represents a single REST route bound to a DB object.
type Endpoint struct {
    ID            uuid.UUID       `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    CollectionID  uuid.UUID       `json:"collection_id" gorm:"type:uuid;not null;index"`
    Collection    *Collection     `json:"collection,omitempty" gorm:"foreignKey:CollectionID"`
    Name          string          `json:"name" gorm:"type:varchar(255);not null"`
    Description   *string         `json:"description,omitempty" gorm:"type:text"`
    Path          string          `json:"path" gorm:"type:varchar(500);not null"`
    Method        string          `json:"method" gorm:"type:varchar(10);not null;default:'GET'"` // GET, POST, PUT, DELETE
    Version       string          `json:"version" gorm:"type:varchar(20);not null;default:'v1'"`
    IsActive      bool            `json:"is_active" gorm:"default:true"`
    CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`

    // DB binding
    DBType    EndpointType  `json:"db_type" gorm:"type:varchar(50);not null"`
    Schema    string        `json:"schema" gorm:"type:varchar(100);default:'public'"`
    TableName string        `json:"table_name" gorm:"type:varchar(255)"`
    FuncName  string        `json:"func_name" gorm:"type:varchar(255)"`
    Params    []byte        `json:"params,omitempty" gorm:"type:jsonb"` // parameter definitions
    Operations []byte       `json:"operations,omitempty" gorm:"type:jsonb"` // allowed operations

    // Security
    SecurityPolicyJSON []byte `json:"security_policy,omitempty" gorm:"type:jsonb"`

    // Header mapping
    AuthHeader   string `json:"auth_header" gorm:"type:varchar(100);default:'Authorization'"`
    ParamHeaders []byte `json:"param_headers,omitempty" gorm:"type:jsonb"`

    // Body mapping
    BodyMappingJSON []byte `json:"body_mapping,omitempty" gorm:"type:jsonb"`
}

// GetSecurityPolicy deserializes the security policy.
func (e *Endpoint) GetSecurityPolicy() SecurityPolicy {
    var policy SecurityPolicy
    json.Unmarshal(e.SecurityPolicyJSON, &policy)
    return policy
}

// SetSecurityPolicy serializes and stores the security policy.
func (e *Endpoint) SetSecurityPolicy(policy SecurityPolicy) error {
    b, err := json.Marshal(policy)
    if err != nil {
        return err
    }
    e.SecurityPolicyJSON = b
    return nil
}

// Validate checks endpoint field constraints.
func (e *Endpoint) Validate() error {
    return validator.New().Struct(e)
}
```

---

## Step 04.07 — API Log Entity

**Build:** Create `backend/internal/domain/entity/apilog.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// LogLevel represents log severity.
type LogLevel string

const (
    LevelDebug LogLevel = "DEBUG"
    LevelInfo  LogLevel = "INFO"
    LevelWarn  LogLevel = "WARN"
    LevelError LogLevel = "ERROR"
    LevelCritical LogLevel = "CRITICAL"
)

// APILog represents a request log entry.
type APILog struct {
    ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    RequestID   string     `json:"request_id" gorm:"type:uuid;not null;index"`
    WorkspaceID *uuid.UUID `json:"workspace_id,omitempty" gorm:"type:uuid;index"`
    UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
    Method      string     `json:"method" gorm:"type:varchar(10);not null;index"`
    Path        string     `json:"path" gorm:"type:varchar(1000);not null;index"`
    StatusCode  int        `json:"status_code" gorm:"not null;index"`
    LatencyMs   int64      `json:"latency_ms" gorm:"not null"`
    LogLevel    LogLevel   `json:"log_level" gorm:"type:varchar(20);not null"`
    Message     string     `json:"message" gorm:"type:text"`
    Meta        []byte     `json:"meta,omitempty" gorm:"type:jsonb"`
    CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
}

// Validate checks log field constraints.
func (l *APILog) Validate() error {
    return validator.New().Struct(l)
}
```

---

## Step 04.08 — Analytics & Alert Entities

**Build:** Create `backend/internal/domain/entity/analytics.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// AnalyticsMetric represents an aggregated API metric.
type AnalyticsMetric struct {
    ID           uuid.UUID   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    WorkspaceID  uuid.UUID   `json:"workspace_id" gorm:"type:uuid;not null;index"`
    EndpointID   *uuid.UUID  `json:"endpoint_id,omitempty" gorm:"type:uuid;index"`
    MetricName   string      `json:"metric_name" gorm:"type:varchar(100);not null"`
    MetricValue  float64     `json:"metric_value" gorm:"not null"`
    PeriodStart  time.Time   `json:"period_start" gorm:"not null;index"`
    PeriodEnd    time.Time   `json:"period_end" gorm:"not null"`
    Labels       []byte      `json:"labels,omitempty" gorm:"type:jsonb"`
    CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime;not null"`
}

// Validate checks analytics field constraints.
func (a *AnalyticsMetric) Validate() error {
    return validator.New().Struct(a)
}
```

**Build:** Create `backend/internal/domain/entity/alert.go`:

```go
package entity

import (
    "time"

    "github.com/google/uuid"
)

// AlertTrigger defines what condition triggers an alert.
type AlertTrigger string

const (
    TriggerErrorRate  AlertTrigger = "error_rate"
    TriggerLatency    AlertTrigger = "latency"
    TriggerAuthFail   AlertTrigger = "auth_failure_burst"
    TriggerDBDown     AlertTrigger = "db_connection_loss"
    TriggerRateLimit  AlertTrigger = "rate_limit_exceeded"
)

// AlertActionType defines how an alert is delivered.
type AlertActionType string

const (
    ActionWebhook AlertActionType = "webhook"
    ActionEmail   AlertActionType = "email"
    ActionMQTT    AlertActionType = "mqtt"
)

// AlertRule represents an alert configuration.
type AlertRule struct {
    ID            uuid.UUID      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name          string         `json:"name" gorm:"type:varchar(255);not null"`
    WorkspaceID   uuid.UUID      `json:"workspace_id" gorm:"type:uuid;not null;index"`
    EndpointID    *uuid.UUID     `json:"endpoint_id,omitempty" gorm:"type:uuid;index"`
    Trigger       AlertTrigger   `json:"trigger" gorm:"type:varchar(50);not null"`
    Threshold     float64        `json:"threshold" gorm:"not null"`
    WindowMinutes int            `json:"window_minutes" gorm:"not null"`
    Actions       []byte         `json:"actions,omitempty" gorm:"type:jsonb"`
    IsEnabled     bool           `json:"is_enabled" gorm:"default:true"`
    CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
    UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
}

// Validate checks alert rule field constraints.
func (a *AlertRule) Validate() error {
    return validator.New().Struct(a)
}

// AlertEvent represents a fired alert notification.
type AlertEvent struct {
    ID           uuid.UUID     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    RuleID       uuid.UUID     `json:"rule_id" gorm:"type:uuid;not null;index"`
    WorkspaceID  uuid.UUID     `json:"workspace_id" gorm:"type:uuid;not null;index"`
    Trigger      AlertTrigger  `json:"trigger" gorm:"type:varchar(50);not null"`
    CurrentValue float64       `json:"current_value" gorm:"not null"`
    Threshold    float64       `json:"threshold" gorm:"not null"`
    Message      string        `json:"message" gorm:"type:text"`
    Notified     bool          `json:"notified" gorm:"default:false"`
    CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime;not null"`
}
```

---

## Step 04.09 — Compile-Time Interface Checks

**Build:** Add compile-time checks in a new file `backend/internal/domain/entity/checks.go`:

```go
package entity

import "github.com/go-playground/validator/v10"

// Ensure validator package is imported.
var _ = validator.New

// Compile-time interface satisfaction checks are in repository layer.
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add all domain entities with validation and domain methods"
```
