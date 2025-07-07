package handlers

import (
	"auth-api/internal/middleware"
	"auth-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

// GetProfile returns the current user's profile
func GetProfile(c *fiber.Ctx) error {
	userEmail := middleware.GetUserEmail(c)

	// Get user from database
	user, err := models.GetUserByEmail(userEmail)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile retrieved successfully",
		"user": fiber.Map{
			"id":           user.ID,
			"email":        user.Email,
			"is_verified":  user.IsVerified,
			"role":         user.Role,
			"created_at":   user.CreatedAt,
		},
	})
}

// UpdateProfile updates the current user's profile
func UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	userEmail := middleware.GetUserEmail(c)

	// For now, just return a success message
	// You can add actual profile update logic here
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user_id": userID,
		"email":   userEmail,
	})
}

// AdminOnly is an example of an admin-only endpoint
func AdminOnly(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Admin access granted",
		"user_id": userID,
		"role":    userRole,
		"data":    "This is admin-only data",
	})
} 