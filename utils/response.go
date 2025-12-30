package utils

import (
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SuccessResponse(c echo.Context, statusCode int, data interface{}) error {
	return c.JSON(statusCode, data)
}

func ErrorResponseJSON(c echo.Context, statusCode int, code, message string, details interface{}) error {
	return c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func BadRequest(c echo.Context, message string, details interface{}) error {
	return ErrorResponseJSON(c, 400, "BAD_REQUEST", message, details)
}

func Unauthorized(c echo.Context, message string) error {
	return ErrorResponseJSON(c, 401, "UNAUTHORIZED", message, nil)
}

func Forbidden(c echo.Context, message string) error {
	return ErrorResponseJSON(c, 403, "FORBIDDEN", message, nil)
}

func NotFound(c echo.Context, message string) error {
	return ErrorResponseJSON(c, 404, "NOT_FOUND", message, nil)
}

func Conflict(c echo.Context, message string) error {
	return ErrorResponseJSON(c, 409, "CONFLICT", message, nil)
}

func InternalServerError(c echo.Context, message string) error {
	return ErrorResponseJSON(c, 500, "INTERNAL_SERVER_ERROR", message, nil)
}
