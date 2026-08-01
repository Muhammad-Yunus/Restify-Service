package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role represents an authorization role.
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"` // system roles cannot be deleted
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`

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
	return validateStruct(r)
}
