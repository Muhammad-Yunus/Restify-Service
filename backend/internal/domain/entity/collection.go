package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Collection groups related endpoints under a common namespace.
type Collection struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null" validate:"required"`
	Description *string    `json:"description,omitempty" gorm:"type:text"`
	Slug        string     `json:"slug" gorm:"type:varchar(255);not null;index" validate:"required"`
	WorkspaceID uuid.UUID  `json:"workspace_id" gorm:"type:uuid;not null;index" validate:"required"`
	Workspace   *Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime;not null"`

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
	return validateStruct(c)
}
