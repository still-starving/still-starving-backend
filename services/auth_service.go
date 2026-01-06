package services

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/utils"
)

type AuthService struct {
	userRepo               *repository.UserRepository
	redisService           *RedisService
	jwtSecret              string
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, redisService *RedisService, jwtSecret string, accessTokenExp, refreshTokenExp time.Duration) *AuthService {
	return &AuthService{
		userRepo:               userRepo,
		redisService:           redisService,
		jwtSecret:              jwtSecret,
		accessTokenExpiration:  accessTokenExp,
		refreshTokenExpiration: refreshTokenExp,
	}
}

func (s *AuthService) Register(req *models.RegisterRequest) error {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessTokenExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtSecret, s.refreshTokenExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	ctx := context.Background()
	if err := s.redisService.StoreRefreshToken(ctx, user.ID, refreshToken, s.refreshTokenExpiration); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.accessTokenExpiration),
		User:         *user,
	}, nil
}

func (s *AuthService) RefreshTokens(refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check token type
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Check if token exists in Redis
	ctx := context.Background()
	userID, err := s.redisService.GetUserIDByToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// Verify user ID matches
	if userID != claims.UserID {
		return nil, fmt.Errorf("token user mismatch")
	}

	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessTokenExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtSecret, s.refreshTokenExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Revoke old refresh token
	s.redisService.RevokeRefreshToken(ctx, refreshToken)

	// Store new refresh token
	if err := s.redisService.StoreRefreshToken(ctx, user.ID, newRefreshToken, s.refreshTokenExpiration); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.accessTokenExpiration),
		User:         *user,
	}, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	ctx := context.Background()
	return s.redisService.RevokeRefreshToken(ctx, refreshToken)
}

func (s *AuthService) GetCurrentUser(userID string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *AuthService) UpdateUser(userID string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.PreferredRadiusKm != nil {
		user.PreferredRadiusKm = *req.PreferredRadiusKm
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
