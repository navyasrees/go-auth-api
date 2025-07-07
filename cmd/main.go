package main

import (
	internal "auth-api/internal"
	"auth-api/internal/config"
	"auth-api/internal/kafka"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	config.InitDB()

	// Initialize Kafka producer
	kafka.InitProducer()

	// Start email consumer
	kafka.StartEmailConsumer()

	app := fiber.New()

	// Setup all routes
	internal.SetupRoutes(app)

	app.Listen(":3000")
} 