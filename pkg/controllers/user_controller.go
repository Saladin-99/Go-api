package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"Go-api/pkg/database/mongodb/models"
	"Go-api/pkg/database/mongodb/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userRepository *repository.UserRepository
	logger         *log.Logger
}

func NewUserController(logger *log.Logger, userRepository *repository.UserRepository) *UserController {
	return &UserController{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *UserController) SignUp(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request body
	if user.Name == "" || user.Email == "" || user.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name, email, and password are required"})
		return
	}

	// Check if the email is already in use
	existingUser, err1 := c.userRepository.GetUserByEmail(user.Email)
	if err1 != nil {
		// Handle the error (e.g., log it, return an internal server error)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if existingUser != nil {
		// Email is already in use
		ctx.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		return
	}

	err := c.userRepository.CreateUser(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}

func (c *UserController) SignIn(ctx *gin.Context) {
	var signInData struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&signInData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userRepository.AuthenticateUser(signInData.Email, signInData.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	accessToken, refreshToken, err := c.generateJWTToken(user.ID.Hex())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate JWT token"})
		return
	}

	// Return success response with JWT tokens
	ctx.JSON(http.StatusOK, gin.H{
		"message":       "user authenticated successfully",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// generateJWTToken generates JWT access and refresh tokens
func (c *UserController) generateJWTToken(userID string) (string, string, error) {
	// Define JWT claims
	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), // Access token expires in 15 minutes
	}
	refreshTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token expires in 7 days
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", "", err
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (c *UserController) RefreshToken(ctx *gin.Context) {
	var refreshTokenData struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&refreshTokenData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the refresh token and extract user ID
	userID, err := c.validateRefreshToken(refreshTokenData.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Generate a new access token
	accessToken, _, err := c.generateJWTToken(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate JWT token"})
		return
	}

	// Return the new access token
	ctx.JSON(http.StatusOK, gin.H{
		"message":      "access token refreshed successfully",
		"access_token": accessToken,
	})
}

// validateRefreshToken validates the refresh token and returns the user ID
func (c *UserController) validateRefreshToken(refreshToken string) (string, error) {
	// Parse and validate the refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key used for signing
		return []byte("your-secret-key"), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	// Extract user ID from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid user ID")
	}

	return userID, nil
}
