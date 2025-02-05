# Noti RabbitMQ Service


## Description

The RabbitMQ Notification Service is a scalable notification service written in Go that integrates with RabbitMQ for message handling and notification distribution. It is designed to provide a reliable solution for delivering notifications (e.g., alerts, warnings, acknowledgements) across your applications. The service listens for messages on a RabbitMQ queue and triggers real-time notifications based on the content received.

## Requirements

- Go 1.19 or higher
- RabbitMQ 3.x (Ensure RabbitMQ is up and running)
- Docker (for running the docker compose)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/marcusnagy/noti-rabbitmq-service.git
   cd noti-rabbitmq-service
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Build the service:
   ```bash
   go build -o noti-rabbitmq-service
   ```

## Configuration

The service requires the following environment variables to be set:

- `PORTS_HTTP_PORT`: The port on which the service will listen for HTTP requests.
- `PORTS_HEALTH_PORT`: The port on which the service will listen for health checks.
- `RABBITMQ_URL`: The URL of the RabbitMQ server.

Example configuration:
```bash
export PORTS_HTTP_PORT=8080
export PORTS_HEALTH_PORT=8081
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

## Usage

1. Ensure RabbitMQ is running and accessible in your environment. By running the docker compose.
```bash
make docker-up
```

2. Run the backend service:
```bash
./noti-rabbitmq-service
```

4. API Endpoints:
   - **Health Check**: `GET /health`
     - Returns the service health status.
   - **Create Notification**: `POST /v1/notifications`
     - Publishes a new notification to RabbitMQ. The payload must be in JSON format.
   - **List Notification Types**: `GET /v1/notifications/types`
     - Retrieves supported notification types.
   - **List Applications**: `GET /v1/applications`
     - Retrieves supported applications.

4. Swagger Documentation:
   The service comes with integrated Swagger documentation. To generate and view them:
   - Install swag if not already installed:
     ```bash
     go install github.com/swaggo/swag/cmd/swag@latest
     ```
   - Generate the Swagger docs:
     ```bash
     swag init -g cmd/backend/main.go --parseDependency --parseInternal
     ```
   - Run the service and navigate to: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)


## Contributing

Contributions are welcome! Please fork the repository and submit pull requests with improvements and bug fixes.

