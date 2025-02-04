package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/marcus.jonathan/noti-rabbitmq/internal/rabbitmq"
)

func simulateCar() {
	log.Printf("Starting car simulation...")

	// Connect to RabbitMQ
	client, err := rabbitmq.NewRabbitMQClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer client.Close()

	log.Printf("Connected to RabbitMQ, setting up notification consumer...")

	// Set up notification handling
	if err := client.ConsumeNotifications(func(msg []byte) error {
		var notification rabbitmq.Notification
		if err := json.Unmarshal(msg, &notification); err != nil {
			log.Printf("Error unmarshaling notification: %v", err)
			return err
		}

		if client.IsPendingRequest(notification.CorrelationID) {
			log.Printf("Received notification - Type: %s, Message: %s", notification.Type, notification.Message)
			client.ClearPendingRequests(notification.CorrelationID)
		} else {
			log.Printf("Received unexpected notification - Type: %s, Message: %s", notification.Type, notification.Message)
		}
		return nil
	}); err != nil {
		log.Fatalf("Failed to set up notification consumer: %v", err)
	}

	log.Printf("Starting simulation loop...")

	// Simulate car movement and battery drain
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial values
	batteryLevel := 100.0 * (0.8 + rand.Float64()*0.2)
	lat := 59.3293 + rand.Float64()*0.1 // Stockholm area
	lon := 18.0686 + rand.Float64()*0.1

	for range ticker.C {
		// Update position
		lat += rand.Float64()*0.001 - 0.0005
		lon += rand.Float64()*0.001 - 0.0005

		// Update battery (drain 1-3% each tick)
		batteryLevel -= rand.Float64()*2 + 1
		if batteryLevel < 0 {
			log.Printf("Battery depleted, recharging to 100%%")
			batteryLevel = 100.0 // "Recharge"
		}

		// Create and send metadata update
		metadata := rabbitmq.CarMetadata{
			Latitude:  lat,
			Longitude: lon,
			Battery:   batteryLevel,
			Timestamp: time.Now(),
		}

		log.Printf("Sending update - Location: (%.4f, %.4f), Battery: %.1f%%",
			metadata.Latitude, metadata.Longitude, metadata.Battery)

		if err := client.PublishMetadataAsync(context.Background(), metadata); err != nil {
			log.Printf("Error sending metadata: %v", err)
		}
	}
}

func main() {
	numCars := 2

	for i := 0; i < numCars; i++ {
		go simulateCar()
	}

	select {}
}
