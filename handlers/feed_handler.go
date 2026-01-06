package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type FeedHandler struct {
	feedService *services.FeedService
}

func NewFeedHandler(feedService *services.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// GetFeed godoc
// @Summary Get combined feed
// @Tags feed
// @Produce json
// @Security BearerAuth
// @Param type query string false "Feed type: all, food, hunger"
// @Success 200 {array} interface{}
// @Router /api/feed [get]
func (h *FeedHandler) GetFeed(c echo.Context) error {
	userID := middleware.GetUserID(c)
	feedType := c.QueryParam("type")

	if feedType == "" {
		feedType = "all"
	}

	var lat, lng, radius float64
	if c.QueryParam("lat") != "" {
		lat, _ = strconv.ParseFloat(c.QueryParam("lat"), 64)
	}
	if c.QueryParam("lng") != "" {
		lng, _ = strconv.ParseFloat(c.QueryParam("lng"), 64)
	}
	if c.QueryParam("radius") != "" {
		radius, _ = strconv.ParseFloat(c.QueryParam("radius"), 64)
		radius = radius * 1000 // Convert km to meters
	}

	feed, err := h.feedService.GetFeed(feedType, userID, lat, lng, radius)
	if err != nil {
		return utils.ErrorResponseJSON(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get feed", err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, feed)
}
