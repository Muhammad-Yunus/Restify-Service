package entity

import (
	"time"

	"github.com/google/uuid"
)

// TeamWorkspaceRole defines the permission level a team has in a workspace.
type TeamWorkspaceRole string

const (
	TeamRoleViewer    TeamWorkspaceRole = "viewer"
	TeamRoleReadWrite TeamWorkspaceRole = "read_write"
	TeamRoleAdmin     TeamWorkspaceRole = "admin"
)

// Team represents a group of users.
type Team struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null" validate:"required"`
	WorkspaceID uuid.UUID  `json:"workspace_id" gorm:"type:uuid;not null;index" validate:"required"`
	Workspace   *Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime;not null"`

	// Relations
	Members []*TeamMember `json:"members,omitempty" gorm:"foreignKey:TeamID"`
}

// Validate checks team field constraints.
func (t *Team) Validate() error {
	return validateStruct(t)
}

// TeamMember represents a user's membership in a team.
type TeamMember struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamID   uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index" validate:"required"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index" validate:"required"`
	Team     *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role     string    `json:"role" gorm:"type:varchar(50);default:'member'"`
	JoinedAt time.Time `json:"joined_at" gorm:"autoCreateTime;not null"`
}

// Validate checks team member field constraints.
func (m *TeamMember) Validate() error {
	return validateStruct(m)
}
