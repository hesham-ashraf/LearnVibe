package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EnrollmentStatus represents the status of a course enrollment
type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
	EnrollmentStatusDropped   EnrollmentStatus = "dropped"
)

// Enrollment represents a student's enrollment in a course
type Enrollment struct {
	ID           uuid.UUID        gorm:"type:uuid;primaryKey" json:"id"
	UserID       uuid.UUID        gorm:"type:uuid;index:idx_enrollment_user_course,unique:true" json:"user_id"
	User         User             gorm:"foreignKey:UserID" json:"user,omitempty"
	CourseID     uuid.UUID        gorm:"type:uuid;index:idx_enrollment_user_course,unique:true" json:"course_id"
	Course       Course           gorm:"foreignKey:CourseID" json:"course,omitempty"
	Status       EnrollmentStatus gorm:"type:varchar(20);default:'active'" json:"status"
	EnrolledAt   time.Time        json:"enrolled_at"
	CompletedAt  *time.Time       json:"completed_at,omitempty"
	LastAccessAt *time.Time       json:"last_access_at,omitempty"
	Progress     float32          gorm:"default:0" json:"progress"
	CreatedAt    time.Time        json:"created_at"
	UpdatedAt    time.Time        json:"updated_at"
}

// BeforeCreate hook to set UUID and enrollment time before creation
func (e *Enrollment) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.EnrolledAt.IsZero() {
		e.EnrolledAt = time.Now()
	}
	return nil
}

// UpdateProgress updates the user's progress in the course
func (e *Enrollment) UpdateProgress(progress float32) {
	e.Progress = progress
	if progress >= 100 {
		now := time.Now()
		e.CompletedAt = &now
		e.Status = EnrollmentStatusCompleted
	}
}

// MarkAsDropped marks the enrollment as dropped
func (e *Enrollment) MarkAsDropped() {
	e.Status = EnrollmentStatusDropped
}

// RecordAccess updates the last access time
func (e *Enrollment) RecordAccess() {
	now := time.Now()
	e.LastAccessAt = &now
}