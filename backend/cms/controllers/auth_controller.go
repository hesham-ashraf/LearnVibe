package controllers

import (
	"context"
	"encoding/json"
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

// AuthController handles authentication-related requests
type AuthController struct {
	db           *gorm.DB
	config       *config.Config
	googleConfig *oauth2.Config
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
func NewAuthController(db *gorm.DB, cfg *config.Config) *AuthController {
	// Configure OAuth2 for Google
	googleConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthController{
		db:           db,
		config:       cfg,
		googleConfig: googleConfig,
	}
}

// GoogleLogin initiates Google OAuth2 login
func (ac *AuthController) GoogleLogin(c *gin.Context) {
	// Generate a random state for CSRF protection
	state := uuid.New().String()
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	// Redirect to Google's OAuth2 consent page
	url := ac.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the callback from Google OAuth2
func (ac *AuthController) GoogleCallback(c *gin.Context) {
	// Verify state to prevent CSRF
	stateCookie, _ := c.Cookie("oauth_state")
	state := c.Query("state")
	if state != stateCookie {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := ac.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Get user info from Google
	client := ac.googleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Google"})
		return
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info from Google"})
		return
	}

	// Find or create user
	var user models.User
	result := ac.db.Where("google_id = ?", userInfo.ID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new user with student role by default
			user = models.User{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				GoogleID: userInfo.ID,
				Role:     models.RoleStudent, // Default role is student
			}
			ac.db.Create(&user)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// Generate JWT
	jwtToken, err := ac.generateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return JWT to client (in production, you might want to redirect to frontend with token)
	c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

// generateJWT creates a new JWT for the user
func (ac *AuthController) generateJWT(user *models.User) (string, error) {
	// Create JWT claims with user ID and role
	claims := jwt.MapClaims{
		"sub":  user.ID.String(),                      // Subject: User ID
		"name": user.Name,                             // User's name
		"role": string(user.Role),                     // User's role for RBAC
		"iat":  time.Now().Unix(),                     // Issued at
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token (string) using the secret key
	tokenString, err := token.SignedString([]byte(ac.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
