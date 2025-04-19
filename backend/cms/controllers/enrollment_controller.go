package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"gorm.io/gorm"
)

// EnrollmentController handles enrollment-related requests
type EnrollmentController struct {
	db *gorm.DB
}

// NewEnrollmentController creates a new enrollment controller
func NewEnrollmentController(db *gorm.DB) *EnrollmentController {
	return &EnrollmentController{db: db}
}

// EnrollInCourse handles course enrollment
func (ec *EnrollmentController) EnrollInCourse(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get course ID from URL
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Check if course exists
	var course models.Course
	result := ec.db.First(&course, courseID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Check if user is already enrolled
	var existingEnrollment models.Enrollment
	result = ec.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&existingEnrollment)
	if result.Error == nil {
		// User is already enrolled
		if existingEnrollment.Status == models.EnrollmentStatusDropped {
			// Reactivate enrollment if it was dropped
			existingEnrollment.Status = models.EnrollmentStatusActive
			existingEnrollment.EnrolledAt = time.Now()
			ec.db.Save(&existingEnrollment)
			c.JSON(http.StatusOK, gin.H{"message": "Course enrollment reactivated", "enrollment": existingEnrollment})
			return
		}

		c.JSON(http.StatusConflict, gin.H{"error": "Already enrolled in this course", "enrollment": existingEnrollment})
		return
	}

	// Create new enrollment
	enrollment := models.Enrollment{
		UserID:   userID.(uuid.UUID),
		CourseID: courseID,
		Status:   models.EnrollmentStatusActive,
	}

	if err := ec.db.Create(&enrollment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enroll in course"})
		return
	}

	c.JSON(http.StatusCreated, enrollment)
}

// GetUserEnrollments lists all courses a user is enrolled in
func (ec *EnrollmentController) GetUserEnrollments(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Pagination parameters
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	status := c.DefaultQuery("status", "")

	// Base query
	query := ec.db.Model(&models.Enrollment{}).Where("user_id = ?", userID).Preload("Course")

	// Filter by status if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Execute query with pagination
	var enrollments []models.Enrollment
	query.Scopes(Paginate(page, pageSize)).Find(&enrollments)

	c.JSON(http.StatusOK, enrollments)
}

// GetCourseEnrollments lists all users enrolled in a course (admin/instructor only)
func (ec *EnrollmentController) GetCourseEnrollments(c *gin.Context) {
	// Get course ID from URL
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Check if course exists
	var course models.Course
	result := ec.db.First(&course, courseID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Verify permissions - only course creator or admin can see enrollments
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	if course.CreatorID != userID.(uuid.UUID) && userRole.(string) != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view enrollments for this course"})
		return
	}

	// Pagination parameters
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	status := c.DefaultQuery("status", "")

	// Base query
	query := ec.db.Model(&models.Enrollment{}).Where("course_id = ?", courseID).Preload("User")

	// Filter by status if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Execute query with pagination
	var enrollments []models.Enrollment
	query.Scopes(Paginate(page, pageSize)).Find(&enrollments)

	c.JSON(http.StatusOK, enrollments)
}

// GetEnrollmentDetails gets details of a specific enrollment
func (ec *EnrollmentController) GetEnrollmentDetails(c *gin.Context) {
	// Get enrollment ID from URL
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}

	// Get enrollment with course and user details
	var enrollment models.Enrollment
	result := ec.db.Preload("Course").Preload("User").First(&enrollment, enrollmentID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enrollment not found"})
		return
	}

	// Verify permissions - only the enrolled user, course creator, or admin can see enrollment details
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	if enrollment.UserID != userID.(uuid.UUID) &&
		enrollment.Course.CreatorID != userID.(uuid.UUID) &&
		userRole.(string) != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this enrollment"})
		return
	}

	// Record access if the enrolled user is viewing
	if enrollment.UserID == userID.(uuid.UUID) {
		enrollment.RecordAccess()
		ec.db.Save(&enrollment)
	}

	c.JSON(http.StatusOK, enrollment)
}

// UpdateEnrollmentProgress updates a user's progress in a course
func (ec *EnrollmentController) UpdateEnrollmentProgress(c *gin.Context) {
	// Get enrollment ID from URL
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get enrollment
	var enrollment models.Enrollment
	result := ec.db.First(&enrollment, enrollmentID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enrollment not found"})
		return
	}

	// Verify ownership - only the enrolled user can update their progress
	if enrollment.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this enrollment"})
		return
	}

	// Parse the progress update from request
	var progressUpdate struct {
		Progress float32 `json:"progress" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&progressUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update progress
	enrollment.UpdateProgress(progressUpdate.Progress)
	ec.db.Save(&enrollment)

	c.JSON(http.StatusOK, enrollment)
}

// DropEnrollment allows a user to drop a course
func (ec *EnrollmentController) DropEnrollment(c *gin.Context) {
	// Get enrollment ID from URL
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enrollment ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get enrollment
	var enrollment models.Enrollment
	result := ec.db.First(&enrollment, enrollmentID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enrollment not found"})
		return
	}

	// Verify ownership - only the enrolled user can drop the course
	if enrollment.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to drop this enrollment"})
		return
	}

	// Mark as dropped
	enrollment.MarkAsDropped()
	ec.db.Save(&enrollment)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully dropped the course", "enrollment": enrollment})
}
