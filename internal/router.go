package internal

import (
	"auth-api/internal/handlers"
	"auth-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	auth := app.Group("/auth")
	auth.Post("/signup", handlers.SignUp)
	auth.Post("/verify-user", handlers.VerifyUser)
	auth.Post("/login", handlers.Login)
	auth.Post("/forgot-password", handlers.ForgotPassword)
	auth.Post("/reset-password", handlers.ResetPassword)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)

	// Protected routes - require authentication
	protected := app.Group("/api", middleware.AuthMiddleware())
	protected.Get("/profile", handlers.GetProfile)
	protected.Put("/profile", handlers.UpdateProfile)

	// Admin routes - require admin role
	admin := app.Group("/admin", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	admin.Get("/data", handlers.AdminOnly)
}
