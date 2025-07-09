package middleware

import (
	"auth-api/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT Bearer token and adds user info to context
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token is required",
			})
		}

			// Validate token
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}



	// Add user info to context
	c.Locals("user_id", claims.UserID)
	c.Locals("user_email", claims.Email)
	c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("user_role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		if userRole != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() fiber.Handler {
	return RoleMiddleware("admin")
}

// GetUserID returns user ID from context
func GetUserID(c *fiber.Ctx) string {
	return c.Locals("user_id").(string)
}

// GetUserEmail returns user email from context
func GetUserEmail(c *fiber.Ctx) string {
	return c.Locals("user_email").(string)
}

// GetUserRole returns user role from context
func GetUserRole(c *fiber.Ctx) string {
	return c.Locals("user_role").(string)
} 