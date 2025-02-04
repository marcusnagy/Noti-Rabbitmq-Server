package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/marcus.jonathan/noti-rabbitmq/internal/models"
	"github.com/marcus.jonathan/noti-rabbitmq/internal/rabbitmq"
)

// @Summary List notification types
// @Description Get a list of all available notification types
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns a list of notification types"
// @Router /v1/notifications/types [get]
func ListNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": []models.NotificationType{
			models.NotificationTypeWarning,
			models.NotificationTypeDanger,
			models.NotificationTypeAck,
			models.NotificationTypeError,
			models.NotificationTypeInfo,
		},
	})
}

// @Summary Create a new notification
// @Description Create and publish a new notification to the message broker
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body models.Notification true "Notification object"
// @Success 200 {object} map[string]string "Returns status ok"
// @Failure 400 {object} map[string]string "Bad request - invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/notifications [post]
func CreateNotification(c *gin.Context) {
	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mqServer, ok := c.MustGet("mqServer").(*rabbitmq.RabbitMQServer)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "message broker not available"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	routingKey, ok := c.Copy().MustGet("routingKey").(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "routing key not found"})
	}
	if err := mqServer.PublishNotification(ctx, notification, routingKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
