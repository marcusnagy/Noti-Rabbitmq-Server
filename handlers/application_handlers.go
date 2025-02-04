package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/marcus.jonathan/noti-rabbitmq/internal/models"
)

// @Summary List all applications
// @Description Get a list of all available applications
// @Tags applications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns a list of applications"
// @Router /v1/applications [get]
func ListApplications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": []models.Application{
			models.FourScreen,
			models.CarSync,
			models.ZeekrGPT,
			models.ZeekrPlaces,
			models.Navi,
			models.ZeekrCircle,
		},
	})
}
