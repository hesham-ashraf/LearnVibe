package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"gorm.io/gorm"
)

// CourseController handles course-related requests
type CourseController struct {
	db *gorm.DB
}

// NewCourseController creates a new course controller
func NewCourseController(db *gorm.DB) *CourseController {
	return &CourseController{db: db}
}

// CreateCourse handles course creation
func (cc *CourseController) CreateCourse(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse course data from request
	var course models.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the creator ID
	course.CreatorID = userID.(uuid.UUID)

	// Create the course
	if err := cc.db.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course"})
		return
	}

	c.JSON(http.StatusCreated, course)
}

// GetCourses lists all courses with pagination
func (cc *CourseController) GetCourses(c *gin.Context) {
	var courses []models.Course
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	// Preload creator information but don't include course contents by default
	query := cc.db.Model(&models.Course{}).Preload("Creator")
	query.Scopes(Paginate(page, pageSize)).Find(&courses)

	c.JSON(http.StatusOK, courses)
}

// GetCourse gets a single course by ID
func (cc *CourseController) GetCourse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var course models.Course
	result := cc.db.Preload("Creator").Preload("Contents").First(&course, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	c.JSON(http.StatusOK, course)
}

// UpdateCourse updates a course
func (cc *CourseController) UpdateCourse(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get course ID from URL
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Get existing course
	var course models.Course
	result := cc.db.First(&course, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Check if user is the creator or an admin
	userRole, _ := c.Get("userRole")
	if course.CreatorID != userID.(uuid.UUID) && userRole.(string) != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this course"})
		return
	}

	// Parse update data
	var updateData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update course
	course.Title = updateData.Title
	course.Description = updateData.Description
	cc.db.Save(&course)

	c.JSON(http.StatusOK, course)
}

// DeleteCourse deletes a course
func (cc *CourseController) DeleteCourse(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get course ID from URL
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Get existing course
	var course models.Course
	result := cc.db.First(&course, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Check if user is the creator or an admin
	userRole, _ := c.Get("userRole")
	if course.CreatorID != userID.(uuid.UUID) && userRole.(string) != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this course"})
		return
	}

	// Delete course contents first (cascade delete)
	cc.db.Where("course_id = ?", id).Delete(&models.CourseContent{})

	// Delete the course
	cc.db.Delete(&course)

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

// AddCourseContent adds content to a course
func (cc *CourseController) AddCourseContent(c *gin.Context) {
	// Get user ID from context
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

	// Get existing course
	var course models.Course
	result := cc.db.First(&course, courseID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Check if user is the creator or an admin
	userRole, _ := c.Get("userRole")
	if course.CreatorID != userID.(uuid.UUID) && userRole.(string) != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this course"})
		return
	}

	// Parse content data
	var content models.CourseContent
	if err := c.ShouldBindJSON(&content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set course ID
	content.CourseID = courseID

	// Create the content
	if err := cc.db.Create(&content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add course content"})
		return
	}

	c.JSON(http.StatusCreated, content)
}

// DeleteCourseContent deletes content from a course
func (cc *CourseController) DeleteCourseContent(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get course ID and content ID from URL
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	contentID, err := uuid.Parse(c.Param("contentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	// Get existing course
	var course models.Course
	result := cc.db.First(&course, courseID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Check if user is the creator or an admin
	userRole, _ := c.Get("userRole")
	if course.CreatorID != userID.(uuid.UUID) && userRole.(string) != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this course"})
		return
	}

	// Delete the content
	result = cc.db.Where("id = ? AND course_id = ?", contentID, courseID).Delete(&models.CourseContent{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found or doesn't belong to this course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Content deleted successfully"})
}

// Paginate is a helper function for pagination
func Paginate(page, pageSize string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Implement pagination logic
		// This is a simple example - you'd want to handle conversion errors in production
		var p, ps int
		if p, err := intOrDefault(page, 1); err == nil {
			if p < 1 {
				p = 1
			}
		}
		if ps, err := intOrDefault(pageSize, 10); err == nil {
			if ps < 1 {
				ps = 10
			}
			if ps > 100 {
				ps = 100 // Limit page size to prevent abuse
			}
		}
		offset := (p - 1) * ps
		return db.Offset(offset).Limit(ps)
	}
}

// intOrDefault converts a string to int or returns a default value
func intOrDefault(val string, defaultVal int) (int, error) {
	// This is a stub - you'd implement proper parsing with error handling
	return defaultVal, nil
}
