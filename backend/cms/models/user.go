package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents user roles for RBAC
type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
	RoleAdmin      Role = "admin"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Name      string    `json:"name"`
	GoogleID  string    `gorm:"uniqueIndex" json:"google_id,omitempty"`
	Role      Role      `gorm:"type:varchar(20);default:'student'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to set UUID before user creation
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// HasRole checks if the user has a specified role
func (u *User) HasRole(role Role) bool {
	return u.Role == role
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsInstructor checks if the user is an instructor
func (u *User) IsInstructor() bool {
	return u.Role == RoleInstructor || u.Role == RoleAdmin
}

// IsStudent checks if the user is a student
func (u *User) IsStudent() bool {
	return u.Role == RoleStudent
}
