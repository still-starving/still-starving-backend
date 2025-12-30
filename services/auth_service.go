package services

import (
	"fmt"
	"time"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/utils"
)

type AuthService struct {
	userRepo      *repository.UserRepository
	jwtSecret     string
	jwtExpiration time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpiration time.Duration) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
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

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
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
