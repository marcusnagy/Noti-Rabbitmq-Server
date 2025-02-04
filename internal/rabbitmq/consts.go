package rabbitmq

const (
	// Exchange names
	CarExchange    = "car.events"
	FanoutExchange = "fanout.events"

	// Queue names
	CarMetadataQueue     = "car.metadata"
	CarNotificationQueue = "car.notifications"

	// Routing key patterns
	MetadataRoutingKey     = "car.metadata.#" // # matches zero or more words
	NotificationRoutingKey = "car.notification.#"
)
