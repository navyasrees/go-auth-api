package main

import (
	internal "auth-api/internal"
	"auth-api/internal/config"
	"auth-api/internal/kafka"
	"auth-api/internal/metrics"
	"log"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Register Prometheus metrics
	metrics.RegisterMetrics()

	app := fiber.New()

	// Add Prometheus middleware to collect HTTP metrics
	app.Use(metrics.PrometheusMiddleware())

	// Setup all routes
	internal.SetupRoutes(app)

	// Add Prometheus metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Listen(":3000")
} 