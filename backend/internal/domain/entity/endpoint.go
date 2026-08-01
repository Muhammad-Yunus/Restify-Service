package entity

import (
	"encoding/json"
	"fmt"
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
	OpSelect OperationType = "select"
	OpInsert OperationType = "insert"
	OpUpdate OperationType = "update"
	OpDelete OperationType = "delete"
	OpCustom OperationType = "custom"
)

// SecurityPolicy defines access control for an endpoint.
type SecurityPolicy struct {
	AuthRequired bool     `json:"auth_required"`
	AllowedRoles []string `json:"allowed_roles,omitempty"`
	RateLimit    *int     `json:"rate_limit,omitempty"` // requests per minute
}

// Endpoint represents a single REST route bound to a DB object.
type Endpoint struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CollectionID uuid.UUID   `json:"collection_id" gorm:"type:uuid;not null;index" validate:"required"`
	Collection   *Collection `json:"collection,omitempty" gorm:"foreignKey:CollectionID"`
	Name         string      `json:"name" gorm:"type:varchar(255);not null" validate:"required"`
	Description  *string     `json:"description,omitempty" gorm:"type:text"`
	Path         string      `json:"path" gorm:"type:varchar(500);not null" validate:"required"`
	Method       string      `json:"method" gorm:"type:varchar(10);not null;default:'GET'" validate:"required"` // GET, POST, PUT, DELETE
	Version      string      `json:"version" gorm:"type:varchar(20);not null;default:'v1'" validate:"required"`
	IsActive     bool        `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt    time.Time   `json:"updated_at" gorm:"autoUpdateTime;not null"`

	// DB binding
	DBType     EndpointType `json:"db_type" gorm:"type:varchar(50);not null" validate:"required"`
	Schema     string       `json:"schema" gorm:"type:varchar(100);default:'public'"`
	TableName  string       `json:"table_name" gorm:"type:varchar(255)"`
	FuncName   string       `json:"func_name" gorm:"type:varchar(255)"`
	Params     []byte       `json:"params,omitempty" gorm:"type:jsonb"`     // parameter definitions
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

	if err := json.Unmarshal(e.SecurityPolicyJSON, &policy); err != nil {
		return SecurityPolicy{}
	}

	return policy
}

// SetSecurityPolicy serializes and stores the security policy.
func (e *Endpoint) SetSecurityPolicy(policy SecurityPolicy) error {
	b, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshal security policy: %w", err)
	}

	e.SecurityPolicyJSON = b

	return nil
}

// Validate checks endpoint field constraints.
func (e *Endpoint) Validate() error {
	return validateStruct(e)
}
