# RabbitMQ Notification Service

## Description

A notification service written in Go that integrates with RabbitMQ for message handling and notification distribution. This service provides a reliable and scalable solution for managing and delivering notifications across your applications.

## Requirements

- Go 1.19 or higher
- RabbitMQ 3.x
- Docker (optional, for containerized deployment)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/noti-rabbitmq.git
cd noti-rabbitmq
```

2. Install dependencies:
```bash
go mod download
```

3. Build the service:
```bash
go build -o noti-rabbitmq
```

## Usage

1. Ensure RabbitMQ is running and accessible.

2. Configure the service using environment variables:
```bash
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

3. Run the service:
```bash
./noti-rabbitmq
```

4. Send messages to the configured RabbitMQ queue to trigger notifications.

For detailed API documentation and examples, please refer to the [documentation](docs/README.md).

