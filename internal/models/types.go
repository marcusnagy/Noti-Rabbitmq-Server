package models

import "time"

type NotificationType string
type Application string

const (
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeDanger  NotificationType = "danger"
	NotificationTypeAck     NotificationType = "ack"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeInfo    NotificationType = "info"
)

const (
	FourScreen  Application = "4screen"
	CarSync     Application = "carsync"
	ZeekrGPT    Application = "ZeekrGPT"
	ZeekrPlaces Application = "Zeekr Places"
	Navi        Application = "Navi"
	ZeekrCircle Application = "Zeekr Circle"
)

type Notification struct {
	Application string           `json:"application"`
	Type        NotificationType `json:"type"`
	Message     string           `json:"message"`
	Timestamp   time.Time        `json:"timestamp,omitempty"`
}
