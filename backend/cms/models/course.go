package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentType represents the type of course content
type ContentType string

const (
	ContentTypePDF   ContentType = "pdf"
	ContentTypeVideo ContentType = "video"
	ContentTypeLink  ContentType = "link"
	ContentTypeText  ContentType = "text"
)

// Course represents a course in the system
type Course struct {
	ID          int             `gorm:"primaryKey" json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	CreatorID   uuid.UUID       `gorm:"type:uuid" json:"creator_id"`
	Creator     User            `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Contents    []CourseContent `json:"contents,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// BeforeCreate hook to set UUID before course creation
func (c *Course) BeforeCreate(tx *gorm.DB) error {
	// ID is auto-incremented
	return nil
}

// CourseContent represents content attached to a course
type CourseContent struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	CourseID    int         `gorm:"index" json:"course_id"`
	Course      Course      `gorm:"foreignKey:CourseID" json:"-"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Type        ContentType `gorm:"type:varchar(10)" json:"type"`
	URL         string      `json:"url"`
	Order       int         `json:"order"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// BeforeCreate hook to set UUID before content creation
func (c *CourseContent) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
