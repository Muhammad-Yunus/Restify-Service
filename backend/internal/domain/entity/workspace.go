package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Workspace represents a top-level isolation container.
type Workspace struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null" validate:"required"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	Slug        string    `json:"slug" gorm:"type:varchar(255);uniqueIndex;not null" validate:"required"`
	OwnerID     uuid.UUID `json:"owner_id" gorm:"type:uuid;not null;index" validate:"required"`
	Owner       *User     `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`

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
	return validateStruct(w)
}
