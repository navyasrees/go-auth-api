package handlers

import (
	"auth-api/internal/kafka"
	"auth-api/internal/metrics"
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
		metrics.RecordAuthSignup(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		metrics.RecordAuthSignup(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	err = models.CreateUser(req.Email, hashedPassword)
	if err != nil {
		metrics.RecordAuthSignup(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	otp := utils.GenerateOTP()
	err = models.StoreOTP(req.Email, otp)
	if err != nil {
		metrics.RecordAuthSignup(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store OTP",
		})
	}

	err = kafka.SendOTPEmail(req.Email, otp)
	if err != nil {
		metrics.RecordAuthSignup(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send OTP email",
		})
	}

	metrics.RecordAuthSignup(true)
	metrics.RecordEmailSent("verification", true)
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
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Debug: Log the parsed request
	log.Printf("Parsed request: %+v", req)

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Check if user exists and password is correct
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Check if user is verified
	if !user.IsVerified {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Please verify your email before logging in",
		})
	}

	// Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
		})
	}

	// Record JWT token generation metrics
	metrics.RecordJWTTokenGenerated("access_token")
	metrics.RecordJWTTokenGenerated("refresh_token")

	// Hash and store refresh token
	refreshTokenHash := utils.HashRefreshToken(refreshToken)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	
	err = models.StoreRefreshToken(user.ID, refreshTokenHash, refreshExpiresAt)
	if err != nil {
		metrics.RecordAuthLogin(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store refresh token",
		})
	}

	metrics.RecordAuthLogin(true)
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

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Validate required fields
	if req.Email == "" {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Check if user exists
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		// For security reasons, don't reveal if user exists or not
		// Return success even if user doesn't exist
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "If the email exists in our system, you will receive a password reset OTP",
		})
	}

	// Check if user is verified
	if !user.IsVerified {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please verify your email before requesting password reset",
		})
	}

	// Generate OTP for password reset
	otp := utils.GenerateOTP()
	
	// Store OTP with a different purpose (password reset)
	err = models.StorePasswordResetOTP(req.Email, otp)
	if err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store password reset OTP",
		})
	}

	// Send OTP via email
	err = kafka.SendPasswordResetOTPEmail(req.Email, otp)
	if err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send password reset OTP email",
		})
	}

	metrics.RecordAuthPasswordReset(true)
	metrics.RecordEmailSent("password_reset", true)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists in our system, you will receive a password reset OTP",
	})
}

func ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Validate required fields
	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, OTP, and new password are required",
		})
	}

	// Validate password strength (minimum 6 characters)
	if len(req.NewPassword) < 6 {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters long",
		})
	}

	// Check if user exists
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email or OTP",
		})
	}

	// Check if user is verified
	if !user.IsVerified {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please verify your email before resetting password",
		})
	}

	// Verify password reset OTP
	err = models.VerifyPasswordResetOTP(req.Email, req.OTP)
	if err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	// Update user's password
	err = models.UpdateUserPassword(req.Email, hashedPassword)
	if err != nil {
		metrics.RecordAuthPasswordReset(false)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	// Invalidate all existing refresh tokens for this user (force logout from all devices)
	err = models.DeleteAllRefreshTokensForUser(user.ID)
	if err != nil {
		// Log the error but don't fail the request
		log.Printf("Failed to delete refresh tokens: %v", err)
	}

	metrics.RecordAuthPasswordReset(true)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password reset successfully. Please login with your new password.",
	})
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
