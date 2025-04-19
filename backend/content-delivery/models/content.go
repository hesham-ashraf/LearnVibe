package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentType represents the type of content
type ContentType string

const (
	ContentTypeVideo    ContentType = "video"
	ContentTypeDocument ContentType = "document"
	ContentTypeImage    ContentType = "image"
	ContentTypeAudio    ContentType = "audio"
	ContentTypeArchive  ContentType = "archive"
	ContentTypeOther    ContentType = "other"
)

// Content represents a media content item in the system
type Content struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	CourseID    uuid.UUID   `gorm:"type:uuid;index" json:"course_id"`
	UploadedBy  uuid.UUID   `gorm:"type:uuid;index" json:"uploaded_by"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	FileName    string      `json:"file_name"`
	FilePath    string      `json:"file_path"`
	FileSize    int64       `json:"file_size"`
	FileType    ContentType `gorm:"type:varchar(20)" json:"file_type"`
	MimeType    string      `json:"mime_type"`
	Duration    *int        `json:"duration,omitempty"` // For video/audio in seconds
	IsPublic    bool        `gorm:"default:false" json:"is_public"`
	Downloads   int         `gorm:"default:0" json:"downloads"`
	Views       int         `gorm:"default:0" json:"views"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// BeforeCreate hook to set UUID before content creation
func (c *Content) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IncrementViews increments the view count for the content
func (c *Content) IncrementViews() {
	c.Views++
}

// IncrementDownloads increments the download count for the content
func (c *Content) IncrementDownloads() {
	c.Downloads++
}

// GetPresignedURL is a placeholder for getting a presigned URL for the content
// In a real implementation, this would interact with MinIO/S3
func (c *Content) GetPresignedURL(expiryMinutes int) (string, error) {
	// This would be implemented to get a presigned URL from MinIO/S3
	// For now, it returns a mock URL based on the file path
	return "/api/content/download/" + c.ID.String(), nil
}
