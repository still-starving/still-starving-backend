package routes

import (
	"database/sql"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/yourusername/food-sharing-backend/config"
	"github.com/yourusername/food-sharing-backend/handlers"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/services"
)

func SetupRoutes(e *echo.Echo, db *sql.DB, minioClient *minio.Client, cfg *config.Config) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	foodPostRepo := repository.NewFoodPostRepository(db)
	hungerBroadcastRepo := repository.NewHungerBroadcastRepository(db)
	foodRequestRepo := repository.NewFoodRequestRepository(db)
	hungerOfferRepo := repository.NewHungerOfferRepository(db)

	// Initialize services
	jwtExpiration, _ := time.ParseDuration(cfg.JWT.Expiration)
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, jwtExpiration)
	imageService := services.NewImageService(minioClient, &cfg.MinIO, cfg.Upload.MaxSize)
	foodPostService := services.NewFoodPostService(foodPostRepo, imageService)
	hungerBroadcastService := services.NewHungerBroadcastService(hungerBroadcastRepo)
	feedService := services.NewFeedService(foodPostRepo, hungerBroadcastRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	foodPostHandler := handlers.NewFoodPostHandler(foodPostService, foodRequestRepo)
	hungerBroadcastHandler := handlers.NewHungerBroadcastHandler(hungerBroadcastService, hungerOfferRepo)
	feedHandler := handlers.NewFeedHandler(feedService)
	userHandler := handlers.NewUserHandler(userRepo, foodRequestRepo, foodPostService, hungerBroadcastService)

	// API group
	api := e.Group("/api")

	// Auth routes (no authentication required)
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.GET("/me", authHandler.GetMe, middleware.AuthMiddleware(cfg.JWT.Secret))

	// Feed routes (authentication required)
	api.GET("/feed", feedHandler.GetFeed, middleware.AuthMiddleware(cfg.JWT.Secret))

	// Food post routes
	foodPosts := api.Group("/food-posts")
	foodPosts.POST("", foodPostHandler.CreateFoodPost, middleware.AuthMiddleware(cfg.JWT.Secret))
	foodPosts.GET("", foodPostHandler.GetFoodPosts)
	foodPosts.GET("/:id", foodPostHandler.GetFoodPost)
	foodPosts.PUT("/:id", foodPostHandler.UpdateFoodPost, middleware.AuthMiddleware(cfg.JWT.Secret))
	foodPosts.DELETE("/:id", foodPostHandler.DeleteFoodPost, middleware.AuthMiddleware(cfg.JWT.Secret))
	foodPosts.POST("/:id/request", foodPostHandler.RequestFood, middleware.AuthMiddleware(cfg.JWT.Secret))
	foodPosts.GET("/:id/requests", foodPostHandler.GetFoodRequests, middleware.AuthMiddleware(cfg.JWT.Secret))
	foodPosts.PUT("/:postId/requests/:requestId/accept", foodPostHandler.AcceptRequest, middleware.AuthMiddleware(cfg.JWT.Secret))
	foodPosts.PUT("/:postId/requests/:requestId/reject", foodPostHandler.RejectRequest, middleware.AuthMiddleware(cfg.JWT.Secret))

	// Hunger broadcast routes
	hungerBroadcasts := api.Group("/hunger-broadcasts")
	hungerBroadcasts.POST("", hungerBroadcastHandler.CreateHungerBroadcast, middleware.AuthMiddleware(cfg.JWT.Secret))
	hungerBroadcasts.GET("", hungerBroadcastHandler.GetHungerBroadcasts)
	hungerBroadcasts.GET("/:id", hungerBroadcastHandler.GetHungerBroadcast)
	hungerBroadcasts.DELETE("/:id", hungerBroadcastHandler.DeleteHungerBroadcast, middleware.AuthMiddleware(cfg.JWT.Secret))
	hungerBroadcasts.POST("/:id/offer", hungerBroadcastHandler.OfferFood, middleware.AuthMiddleware(cfg.JWT.Secret))

	// User routes (all require authentication)
	api.GET("/my-requests", userHandler.GetMyRequests, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/my-posts", userHandler.GetMyPosts, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/my-hunger-broadcasts", userHandler.GetMyHungerBroadcasts, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/profile", userHandler.GetProfile, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.PUT("/profile", userHandler.UpdateProfile, middleware.AuthMiddleware(cfg.JWT.Secret))
}
