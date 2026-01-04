package routes

import (
	"database/sql"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/food-sharing-backend/config"
	"github.com/yourusername/food-sharing-backend/handlers"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/services"
)

func SetupRoutes(e *echo.Echo, db *sql.DB, redisClient *redis.Client, minioClient *minio.Client, cfg *config.Config, hub *services.Hub) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	foodPostRepo := repository.NewFoodPostRepository(db)
	postImageRepo := repository.NewPostImageRepository(db)
	hungerBroadcastRepo := repository.NewHungerBroadcastRepository(db)
	foodRequestRepo := repository.NewFoodRequestRepository(db)
	hungerOfferRepo := repository.NewHungerOfferRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	accessTokenExp, _ := time.ParseDuration(cfg.JWT.AccessTokenExpiration)
	refreshTokenExp, _ := time.ParseDuration(cfg.JWT.RefreshTokenExpiration)
	redisService := services.NewRedisService(redisClient)
	authService := services.NewAuthService(userRepo, redisService, cfg.JWT.Secret, accessTokenExp, refreshTokenExp)
	imageService := services.NewImageService(minioClient, &cfg.MinIO, cfg.Upload.MaxSize)
	foodPostService := services.NewFoodPostService(foodPostRepo, postImageRepo, imageService)
	hungerBroadcastService := services.NewHungerBroadcastService(hungerBroadcastRepo)
	feedService := services.NewFeedService(foodPostRepo, hungerBroadcastRepo)
	conversationService := services.NewConversationService(conversationRepo, foodPostRepo)
	messageService := services.NewMessageService(messageRepo, conversationRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	foodPostHandler := handlers.NewFoodPostHandler(foodPostService, foodRequestRepo, hub)
	hungerBroadcastHandler := handlers.NewHungerBroadcastHandler(hungerBroadcastService, hungerOfferRepo, hub)
	feedHandler := handlers.NewFeedHandler(feedService)
	userHandler := handlers.NewUserHandler(userRepo, foodRequestRepo, foodPostService, hungerBroadcastService)
	conversationHandler := handlers.NewConversationHandler(conversationService)
	messageHandler := handlers.NewMessageHandler(messageService)
	wsHandler := handlers.NewWebSocketHandler(hub, messageService, conversationService, cfg.JWT.Secret)

	// API group
	api := e.Group("/api")

	// WebSocket endpoint (authentication via query param)
	api.GET("/ws", wsHandler.HandleWebSocket)

	// Auth routes (no authentication required)
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)
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

	// Conversation routes (all require authentication)
	conversations := api.Group("/conversations", middleware.AuthMiddleware(cfg.JWT.Secret))
	conversations.POST("", conversationHandler.CreateOrGetConversation)
	conversations.GET("", conversationHandler.GetUserConversations)
	conversations.GET("/:id", conversationHandler.GetConversation)
	conversations.GET("/:id/messages", messageHandler.GetMessages)
	conversations.PUT("/:id/messages/read", messageHandler.MarkAsRead)

	// Message routes (all require authentication)
	api.GET("/messages/unread-count", messageHandler.GetUnreadCount, middleware.AuthMiddleware(cfg.JWT.Secret))

	// User routes (all require authentication)
	api.GET("/my-requests", userHandler.GetMyRequests, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/my-posts", userHandler.GetMyPosts, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/my-hunger-broadcasts", userHandler.GetMyHungerBroadcasts, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.GET("/profile", userHandler.GetProfile, middleware.AuthMiddleware(cfg.JWT.Secret))
	api.PUT("/profile", userHandler.UpdateProfile, middleware.AuthMiddleware(cfg.JWT.Secret))
}
