package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Health Check
// @Description Get the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
