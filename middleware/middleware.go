package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/marcus.jonathan/noti-rabbitmq/internal/rabbitmq"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Set("start", t)
		c.Next()

		latency := time.Since(t)
		status := c.Writer.Status()
		size := c.Writer.Size()

		log.Printf("| %3d | %13v | %15s | %s | %-7s | %d",
			status,
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			size,
		)
	}
}

func SetContext(mqServer *rabbitmq.RabbitMQServer, routingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("mqServer", mqServer)
		c.Set("routingKey", routingKey)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
