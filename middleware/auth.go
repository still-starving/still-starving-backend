package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/utils"
)

func AuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return utils.Unauthorized(c, "Missing authorization header")
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return utils.Unauthorized(c, "Invalid authorization header format")
			}

			token := parts[1]
			claims, err := utils.ValidateToken(token, jwtSecret)
			if err != nil {
				return utils.Unauthorized(c, "Invalid or expired token")
			}

			// Store user ID in context
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	userID, _ := c.Get("userID").(string)
	return userID
}
