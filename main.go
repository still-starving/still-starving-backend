package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/yourusername/food-sharing-backend/config"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/routes"
	"github.com/yourusername/food-sharing-backend/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := config.InitDatabase(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	log.Println("✓ Database connected successfully")

	// Initialize MinIO
	minioClient, err := config.InitMinIO(cfg.MinIO)
	if err != nil {
		log.Fatal("Failed to initialize MinIO:", err)
	}
	log.Println("✓ MinIO connected successfully")

	// Initialize Redis
	redisClient, err := config.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	log.Println("✓ Redis connected successfully")

	// Log startup configuration
	log.Println("========================================")
	log.Printf("Environment: %s", cfg.Env)
	log.Printf("Database: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	log.Printf("MinIO: %s (Public URL: %s)", cfg.MinIO.Endpoint, cfg.MinIO.PublicURL)
	log.Printf("Redis: %s:%s", cfg.Redis.Host, cfg.Redis.Port)
	log.Printf("CORS Allowed Origins: %v", cfg.Server.AllowedOrigins)
	log.Println("========================================")

	// Create Echo instance
	e := echo.New()

	// Middleware - Custom Logger with detailed format
	e.Use(echoMiddleware.LoggerWithConfig(echoMiddleware.LoggerConfig{
		Format:           "${time_rfc3339} | ${status} | ${method} ${uri} | ${latency_human} | ${remote_ip} | ${error}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
	}))
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.SetupCORS(cfg.Server.AllowedOrigins))

	// Custom validator
	e.Validator = utils.NewValidator()

	// Setup routes
	routes.SetupRoutes(e, db, redisClient, minioClient, cfg)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Log all registered routes
	log.Println("========================================")
	log.Println("Registered Routes:")
	log.Println("========================================")
	for _, route := range e.Routes() {
		log.Printf("%-6s %s", route.Method, route.Path)
	}
	log.Println("========================================")

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("✓ Server starting on %s", serverAddr)

	// Graceful shutdown
	go func() {
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
