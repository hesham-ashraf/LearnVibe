package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

// DBInterface defines database operations needed by controllers
type DBInterface interface {
	Where(query interface{}, args ...interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
}

// AuthController handles authentication-related endpoints
type AuthController struct {
	db         DBInterface
	config     *config.Config
	oauthConf  *oauth2.Config
	stateStore map[string]time.Time // Simple in-memory state store
}

// GoogleUserInfo represents the response from Google's userinfo endpoint
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// NewAuthController creates a new authentication controller
func NewAuthController(db DBInterface, cfg *config.Config) *AuthController {
	// Initialize OAuth2 config
	var oauthConf *oauth2.Config
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		oauthConf = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		}
	}

	return &AuthController{
		db:         db,
		config:     cfg,
		oauthConf:  oauthConf,
		stateStore: make(map[string]time.Time),
	}
}

// GoogleLogin initiates the OAuth2 login flow
func (ac *AuthController) GoogleLogin(c *gin.Context) {
	if ac.oauthConf == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth2 not configured"})
		return
	}

	// Generate a random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate random state"})
		return
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state with expiration (10 minutes)
	ac.stateStore[state] = time.Now().Add(10 * time.Minute)

	// Cleanup old states
	ac.cleanupStates()

	// Redirect to Google's OAuth 2.0 server
	url := ac.oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the OAuth2 callback
func (ac *AuthController) GoogleCallback(c *gin.Context) {
	// Get state and code from request
	receivedState := c.Query("state")
	code := c.Query("code")

	// Verify state
	expirationTime, validState := ac.stateStore[receivedState]
	if !validState || time.Now().After(expirationTime) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired state"})
		return
	}

	// Clean up used state
	delete(ac.stateStore, receivedState)

	// Exchange auth code for token
	token, err := ac.oauthConf.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange auth code for token"})
		return
	}

	// Get user info from Google
	client := ac.oauthConf.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	// Parse user info
	var userInfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// Find or create user - simplified database operations
	var user models.User
	db := ac.db.Where("email = ?", userInfo.Email)
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	db = db.First(&user)

	if db.Error == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			ID:        uuid.New(),
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			Role:      models.RoleStudent, // Default role for new users
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		db = ac.db.Create(&user)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate JWT token
	jwtToken, err := ac.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set token in cookie and redirect to frontend
	c.SetCookie(
		"auth_token",
		jwtToken,
		3600*24, // 24 hours
		"/",
		"",
		false, // HTTPS only? Set to true in production
		true,  // HTTP only
	)

	// If there's a return_to parameter, redirect there, otherwise to home
	returnTo := c.Query("return_to")
	if returnTo == "" {
		returnTo = "/"
	}

	c.Redirect(http.StatusTemporaryRedirect, returnTo)
}

// Login handles email/password login
func (ac *AuthController) Login(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var user models.User
	// Simplified database lookup to work better with mocks
	db := ac.db.Where("email = ?", loginRequest.Email)
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	db = db.First(&user)
	if db.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if !user.VerifyPassword(loginRequest.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := ac.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	var registerRequest struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if user already exists - simplified database operations
	var existingUser models.User
	db := ac.db.Where("email = ?", registerRequest.Email)
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	db = db.First(&existingUser)
	if db.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	} else if db.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Create new user
	user := models.User{
		ID:        uuid.New(),
		Name:      registerRequest.Name,
		Email:     registerRequest.Email,
		Role:      models.RoleStudent, // Default role for new users
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set password hash
	if err := user.SetPassword(registerRequest.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	db = ac.db.Create(&user)
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := ac.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

// generateJWT generates a JWT token for a user
func (ac *AuthController) generateJWT(user models.User) (string, error) {
	// Create token with claims
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"name":  user.Name,
		"email": user.Email,
		"role":  string(user.Role),
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(ac.config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// cleanupStates removes expired states from memory
func (ac *AuthController) cleanupStates() {
	now := time.Now()
	for state, expiry := range ac.stateStore {
		if now.After(expiry) {
			delete(ac.stateStore, state)
		}
	}
}

// GetCurrentUser returns the currently authenticated user
func (ac *AuthController) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var user models.User
	db := ac.db.First(&user, "id = ?", userID)
	if db.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}
