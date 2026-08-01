package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
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
	Schema     string
	Name       string
	Columns    []ColumnSchema
	PrimaryKey []string
}

// ColumnSchema represents a column definition.
type ColumnSchema struct {
	Name       string
	Type       string
	IsNullable bool
	IsPrimary  bool
	Default    *string
}

// FunctionInfo represents a discovered database function.
type FunctionInfo struct {
	Schema     string
	Name       string
	Params     []ParamSchema
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
	Name string
	Type string
	Mode string // IN, OUT, INOUT
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
