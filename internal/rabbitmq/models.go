package rabbitmq

import "time"

// CarMetadata represents all car telemetry data
type CarMetadata struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Battery   float64   `json:"battery"`
	Timestamp time.Time `json:"timestamp"`
}

// Notification represents a message to be sent to a car
type Notification struct {
	CorrelationID string    `json:"correlation_id"`
	Type          string    `json:"type"` // "coupon", "poi", "charging_station"
	Message       string    `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
}
