package handlers

import (
	"net/http"

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

	feed, err := h.feedService.GetFeed(feedType, userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get feed")
	}

	return utils.SuccessResponse(c, http.StatusOK, feed)
}
