package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/models"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/services"
)

// ContentController handles content-related requests
type ContentController struct {
	db      *models.Database
	storage *services.StorageService
}

// NewContentController creates a new content controller
func NewContentController(db *models.Database, storage *services.StorageService) *ContentController {
	return &ContentController{
		db:      db,
		storage: storage,
	}
}

// UploadContent handles file uploads
func (cc *ContentController) UploadContent(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get course ID from form data
	courseIDStr := c.PostForm("course_id")
	if courseIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Course ID is required"})
		return
	}

	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Parse form data
	title := c.PostForm("title")
	description := c.PostForm("description")
	isPublicStr := c.PostForm("is_public")
	isPublic := isPublicStr == "true"

	// Ensure title is provided
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	// Get file from form
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Get file info
	fileSize := fileHeader.Size
	fileName := fileHeader.Filename
	fileExt := filepath.Ext(fileName)

	// Determine content type based on file extension
	contentType := models.ContentTypeOther
	switch fileExt {
	case ".mp4", ".avi", ".mov", ".wmv":
		contentType = models.ContentTypeVideo
	case ".pdf", ".doc", ".docx", ".txt", ".ppt", ".pptx":
		contentType = models.ContentTypeDocument
	case ".jpg", ".jpeg", ".png", ".gif", ".svg":
		contentType = models.ContentTypeImage
	case ".mp3", ".wav", ".ogg":
		contentType = models.ContentTypeAudio
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		contentType = models.ContentTypeArchive
	}

	// Create a new content record
	contentID := uuid.New()
	content := models.Content{
		ID:          contentID,
		CourseID:    courseID,
		UploadedBy:  userID.(uuid.UUID),
		Title:       title,
		Description: description,
		FileName:    fileName,
		FilePath:    fmt.Sprintf("%s/%s%s", courseID, contentID, fileExt),
		FileSize:    fileSize,
		FileType:    contentType,
		MimeType:    fileHeader.Header.Get("Content-Type"),
		IsPublic:    isPublic,
	}

	// Save content metadata to database
	if err := cc.db.DB.Create(&content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save content metadata"})
		return
	}

	// Upload file to storage
	ctx := c.Request.Context()
	err = cc.storage.UploadFile(ctx, content.FilePath, file, fileSize, content.MimeType)
	if err != nil {
		// Rollback database entry if storage upload fails
		cc.db.DB.Delete(&content)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, content)
}

// GetContent retrieves content metadata
func (cc *ContentController) GetContent(c *gin.Context) {
	// Get content ID from URL
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	// Get content from database
	var content models.Content
	result := cc.db.DB.First(&content, contentID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
		return
	}

	// Check permissions (if not public)
	if !content.IsPublic {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required for non-public content"})
			return
		}

		// TODO: Add permission check to verify user has access to this course
	}

	// Increment view count
	content.IncrementViews()
	cc.db.DB.Save(&content)

	// Cache content metadata in Redis
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("content:%s", contentID)
	cc.db.Redis.Set(ctx, cacheKey, content, 30*time.Minute)

	c.JSON(http.StatusOK, content)
}

// GetContentDownloadURL generates a download URL for content
func (cc *ContentController) GetContentDownloadURL(c *gin.Context) {
	// Get content ID from URL
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	// Try to get from cache first
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("content:%s", contentID)
	var content models.Content
	_, err = cc.db.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Use cached content if available
		// In a real implementation, we would unmarshal the JSON
		// For now, just get from DB
		result := cc.db.DB.First(&content, contentID)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	} else {
		// Get content from database
		result := cc.db.DB.First(&content, contentID)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	}

	// Check permissions (if not public)
	if !content.IsPublic {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required for non-public content"})
			return
		}

		// TODO: Add permission check to verify user has access to this course
	}

	// Parse expiry time from query (default 60 minutes)
	expiryStr := c.DefaultQuery("expiry", "60")
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil || expiry < 1 || expiry > 1440 { // 1 min to 24 hours
		expiry = 60
	}

	// Generate presigned URL
	url, err := cc.storage.GetPresignedURL(ctx, content.FilePath, time.Duration(expiry)*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	// Increment download count
	content.IncrementDownloads()
	cc.db.DB.Save(&content)

	c.JSON(http.StatusOK, gin.H{
		"url":     url,
		"expires": fmt.Sprintf("%d minutes", expiry),
	})
}

// DeleteContent deletes content
func (cc *ContentController) DeleteContent(c *gin.Context) {
	// Get content ID from URL
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	// Get content from database
	var content models.Content
	result := cc.db.DB.First(&content, contentID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
		return
	}

	// Check permissions
	userID, exists := c.Get("userID")
	userRole, _ := c.Get("userRole")
	if !exists || (content.UploadedBy != userID.(uuid.UUID) && userRole != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this content"})
		return
	}

	// Delete file from storage
	ctx := c.Request.Context()
	err = cc.storage.DeleteFile(ctx, content.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from storage"})
		return
	}

	// Delete content from database
	cc.db.DB.Delete(&content)

	// Delete from cache
	cacheKey := fmt.Sprintf("content:%s", contentID)
	cc.db.Redis.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, gin.H{"message": "Content deleted successfully"})
}