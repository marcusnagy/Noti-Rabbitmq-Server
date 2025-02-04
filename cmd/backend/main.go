package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/marcus.jonathan/noti-rabbitmq/config"
	_ "github.com/marcus.jonathan/noti-rabbitmq/docs" // This is required for swagger
	"github.com/marcus.jonathan/noti-rabbitmq/handlers"
	"github.com/marcus.jonathan/noti-rabbitmq/internal/rabbitmq"
	"github.com/marcus.jonathan/noti-rabbitmq/middleware"
)

// @title Notification Service API
// @version 1.0
// @description This is a notification service that handles car metadata and sends notifications.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// NotificationState holds the last time each notification type was sent
// per car (using CarID).
type NotificationState struct {
	LastPOINotification             time.Time
	LastLowBatteryNotification      time.Time
	LastCriticalBatteryNotification time.Time
}

// carStates stores the state per car.
// In a distributed system, this might be in a cache like Redis.
var (
	carStates  = make(map[string]*NotificationState)
	stateMutex sync.Mutex
)

// getCarState fetches or initializes the NotificationState for a given car.
func getCarState(carID string) *NotificationState {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	if state, exists := carStates[carID]; exists {
		return state
	}
	newState := &NotificationState{}
	carStates[carID] = newState
	return newState
}

func main() {
	log.Printf("Starting backend service...")

	// Connect to RabbitMQ
	server, err := rabbitmq.NewRabbitMQServer("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer server.Close()

	config := config.LoadConfig()
	r := gin.Default()

	r.Use(
		gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
			)
		}),
		middleware.CORSMiddleware(),
		middleware.Logger(),
		middleware.SetContext(server, ""),
	)

	r.GET("/health", handlers.HealthCheck)

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/v1")
	{
		// Endpoint for creating a notification at /v1/notifications
		v1.POST("/notifications", handlers.CreateNotification)

		// Nested group so the GET endpoint resolves to /v1/notifications/types/list
		notificationsGroup := v1.Group("/notifications")
		{
			notificationsGroup.GET("/types", handlers.ListNotifications)
		}
		applicationsGroup := v1.Group("/applications")
		{
			applicationsGroup.GET("", handlers.ListApplications)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Ports.HTTPPort),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Connected to RabbitMQ, setting up consumers...")

	// Handle car metadata updates
	if err := server.ConsumeRPC(
		rabbitmq.CarMetadataQueue,
		rabbitmq.MetadataRoutingKey,
		func(delivery amqp.Delivery) (<-chan rabbitmq.Notification, error) {
			notifChan := make(chan rabbitmq.Notification)

			go func() {
				defer close(notifChan)

				// Unmarshal the message payload into metadata.
				// We assume that CarMetadata now includes a CarID field.
				var metadata rabbitmq.CarMetadata
				if err := json.Unmarshal(delivery.Body, &metadata); err != nil {
					log.Printf("Error unmarshaling metadata: %v", err)
					notifChan <- rabbitmq.Notification{
						CorrelationID: delivery.CorrelationId,
						Type:          "ack",
						Message:       "Invalid metadata format",
					}
					return
				}
				log.Printf("Received metadata update for CarID: %s, Location: (%.4f, %.4f), Battery: %.1f%%",
					delivery.ReplyTo, metadata.Latitude, metadata.Longitude, metadata.Battery)

				// Use per-car state to avoid spamming notifications.
				state := getCarState(delivery.ReplyTo)
				now := time.Now()
				// Define a minimum interval for the same notification type.
				const notificationInterval = 10 * time.Minute

				// --- POI Notification ---
				// For example, if the car is passing a certain latitude (e.g. near a café)...
				if metadata.Latitude > 59.35 {
					if now.Sub(state.LastPOINotification) > notificationInterval {
						notifChan <- rabbitmq.Notification{
							Type:      "poi",
							Message:   "There's a nice café nearby!",
							Timestamp: now,
						}
						state.LastPOINotification = now
					}
				} else {
					// If the car leaves the area, you could choose to reset the state:
					state.LastPOINotification = time.Time{}
				}

				// --- Battery Notifications ---
				// Send a critical battery warning and suppress subsequent messages for a while.
				if metadata.Battery < 10.0 {
					if now.Sub(state.LastCriticalBatteryNotification) > notificationInterval {
						notifChan <- rabbitmq.Notification{
							Type:      "warning",
							Message:   "Critical battery level! Please charge immediately!",
							Timestamp: now,
						}
						state.LastCriticalBatteryNotification = now
					}
				} else if metadata.Battery < 50.0 {
					if now.Sub(state.LastLowBatteryNotification) > notificationInterval {
						notifChan <- rabbitmq.Notification{
							Type:      "warning",
							Message:   "Low battery! Nearest charging station is 2km away.",
							Timestamp: now,
						}
						state.LastLowBatteryNotification = now
					}
				} else {
					// Reset battery warning states when the battery recovers.
					state.LastLowBatteryNotification = time.Time{}
					state.LastCriticalBatteryNotification = time.Time{}
				}

				// Always send an ACK so that the metadata update is marked as processed.
				notifChan <- rabbitmq.Notification{
					CorrelationID: delivery.CorrelationId,
					Timestamp:     time.Now(),
					Type:          "ack",
					Message:       "Metadata processed",
				}
			}()

			return notifChan, nil
		},
	); err != nil {
		log.Fatalf("Failed to set up metadata consumer: %v", err)
	}

	log.Printf("Backend service is running. Press Ctrl+C to exit.")

	// Keep the application running.
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	<-ctx.Done()
	log.Println("timeout of 5 seconds.")
	log.Println("Server exiting")
}
