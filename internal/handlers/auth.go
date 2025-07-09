package handlers

import (
	"auth-api/internal/kafka"
	"auth-api/internal/models"
	"auth-api/internal/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyUserRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignUp(c *fiber.Ctx) error {
	var req SignUpRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	err = models.CreateUser(req.Email, hashedPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	otp := utils.GenerateOTP()
	err = models.StoreOTP(req.Email, otp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store OTP",
		})
	}

	err = kafka.SendOTPEmail(req.Email, otp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send OTP email",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
	})
}

func VerifyUser(c *fiber.Ctx) error {
	var req VerifyUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Validate required fields
	if req.Email == "" || req.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and OTP are required",
		})
	}

	err := models.VerifyUserOTP(req.Email, req.OTP)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User verified successfully",
	})
}

func Login(c *fiber.Ctx) error {
	var req LoginRequest

	// Debug: Log the raw body
	body := c.Body()
	log.Printf("Raw request body: %s", string(body))

	if err := c.BodyParser(&req); err != nil {
		log.Printf("BodyParser error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Debug: Log the parsed request
	log.Printf("Parsed request: %+v", req)

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Check if user exists and password is correct
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Check if user is verified
	if !user.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Please verify your email before logging in",
		})
	}

	// Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
		})
	}

	// Hash and store refresh token
	refreshTokenHash := utils.HashRefreshToken(refreshToken)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	
	err = models.StoreRefreshToken(user.ID, refreshTokenHash, refreshExpiresAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store refresh token",
		})
	}



	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    900, // 15 minutes in seconds
		"user": fiber.Map{
			"id":           user.ID,
			"email":        user.Email,
			"is_verified":  user.IsVerified,
			"role":         user.Role,
			"created_at":   user.CreatedAt,
		},
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	// Hash the provided token and check if it exists in database
	tokenHash := utils.HashRefreshToken(req.RefreshToken)
	storedToken, err := models.GetRefreshTokenByHash(tokenHash)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		// Delete expired token
		models.DeleteRefreshToken(tokenHash)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Refresh token expired",
		})
	}

	// Get user to ensure they still exist and are verified
	user, err := models.GetUserByID(storedToken.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if !user.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not verified",
		})
	}

	// Generate new access and refresh tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
		})
	}

	// Hash and store new refresh token
	newRefreshTokenHash := utils.HashRefreshToken(newRefreshToken)
	newRefreshExpiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	
	err = models.StoreRefreshToken(user.ID, newRefreshTokenHash, newRefreshExpiresAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store new refresh token",
		})
	}

	// Delete the old refresh token
	err = models.DeleteRefreshToken(tokenHash)
	if err != nil {
		// Log the error but don't fail the request
		log.Printf("Failed to delete old refresh token: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Token refreshed successfully",
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    900, // 15 minutes in seconds
	})
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func Logout(c *fiber.Ctx) error {
	var req LogoutRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	// Hash the refresh token and delete it from database
	tokenHash := utils.HashRefreshToken(req.RefreshToken)
	err := models.DeleteRefreshToken(tokenHash)
	if err != nil {
		// Token might not exist, but we still return success for security
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
