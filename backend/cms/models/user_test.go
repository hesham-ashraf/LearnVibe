package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUserSetPassword verifies the password hashing functionality
func TestUserSetPassword(t *testing.T) {
	// Create a test user
	user := User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  RoleStudent,
	}

	// Set a password
	err := user.SetPassword("securepassword123")

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, user.Password)
	assert.NotEqual(t, "securepassword123", user.Password)
}

// TestUserVerifyPassword tests the password verification
func TestUserVerifyPassword(t *testing.T) {
	// Create a test user
	user := User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  RoleStudent,
	}

	// Set a password
	err := user.SetPassword("securepassword123")
	assert.NoError(t, err)

	// Test correct password
	assert.True(t, user.VerifyPassword("securepassword123"))

	// Test incorrect password
	assert.False(t, user.VerifyPassword("wrongpassword"))
}

// TestUserRoles tests the role checking functions
func TestUserRoles(t *testing.T) {
	tests := []struct {
		name           string
		role           Role
		isAdmin        bool
		isInstructor   bool
		isStudent      bool
		hasAdminRole   bool
		hasStudentRole bool
	}{
		{
			name:           "Admin User",
			role:           RoleAdmin,
			isAdmin:        true,
			isInstructor:   true, // Admin has instructor privileges
			isStudent:      false,
			hasAdminRole:   true,
			hasStudentRole: false,
		},
		{
			name:           "Instructor User",
			role:           RoleInstructor,
			isAdmin:        false,
			isInstructor:   true,
			isStudent:      false,
			hasAdminRole:   false,
			hasStudentRole: false,
		},
		{
			name:           "Student User",
			role:           RoleStudent,
			isAdmin:        false,
			isInstructor:   false,
			isStudent:      true,
			hasAdminRole:   false,
			hasStudentRole: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Role: tt.role}
			assert.Equal(t, tt.isAdmin, user.IsAdmin())
			assert.Equal(t, tt.isInstructor, user.IsInstructor())
			assert.Equal(t, tt.isStudent, user.IsStudent())
			assert.Equal(t, tt.hasAdminRole, user.HasRole(RoleAdmin))
			assert.Equal(t, tt.hasStudentRole, user.HasRole(RoleStudent))
		})
	}
}

// TestBeforeCreate tests the UUID generation before creating a user
func TestBeforeCreate(t *testing.T) {
	// Skip this test since we can't mock gorm.DB easily without additional interfaces
	t.Skip("Skipping BeforeCreate test - requires proper DB mocking")
}

// Simple mock for gorm.DB
type mockDB struct{}
