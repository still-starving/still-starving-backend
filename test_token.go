package main

import (
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Test token from your logs
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjMzZmZmI1Ny03YTc4LTQ1YTYtOWVhYy0zZjM3MjkxNjVlM2IiLCJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsInR5cGUiOiJhY2Nlc3MiLCJleHAiOjE3Njc1MzM2MDAsImlhdCI6MTc2NzUzMjcwMH0.ZE0AXyysb0yJcrJMULvpwryWP9YCJUTOJbeLMUuXXNs"

	// Your JWT secret from .env
	secret := "your-secret-key-change-in-production"

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		log.Printf("❌ Token parsing error: %v", err)
		return
	}

	if !token.Valid {
		log.Printf("❌ Token is not valid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("❌ Claims are not MapClaims")
		return
	}

	fmt.Println("✅ Token is valid!")
	fmt.Printf("📋 Claims: %+v\n", claims)

	if sub, ok := claims["sub"].(string); ok {
		fmt.Printf("👤 User ID (sub): %s\n", sub)
	} else {
		fmt.Println("❌ 'sub' claim not found or not a string")
	}
}
