package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ErrNotFound is returned when a record is not found in the database.
var ErrNotFound = errors.New("record not found")

var validate = validator.New()

// User represents an application user.
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null" validate:"required,email"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null" validate:"required"`
	FullName     *string   `json:"full_name,omitempty" gorm:"type:varchar(255)"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`

	// Relations
	Roles []*Role `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

// SetPassword hashes the plaintext password.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
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
	return validateStruct(u)
}

// EmailValid returns whether the email format is valid.
func EmailValid(email string) bool {
	return validate.Var(email, "email") == nil
}
