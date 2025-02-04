# Run the application
run:
	go run cmd/app/main.go

# Start/stop Docker containers
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

# Buf generate
# buf-generate:
# 	@echo "Generating Protobuf code..."
# 	buf generate api

# # Buf remove
# buf-remove:
# 	@echo "Removing Protobuf code..."
# 	rm -rf proto/*

# # Manage Database migrations
# migrate-up:
# 	migrate -path db/migrations -database $(DATABASE_URL) up

# migrate-down:
# 	migrate -path db/migrations -database $(DATABASE_URL) down

# migrate-version:
# 	migrate -path db/migrations -database $(DATABASE_URL) version

# migrate-status:
# 	@echo "Migration status:"
# 	migrate -path db/migrations -database $(DATABASE_URL) version || echo "No migrations applied yet."

# # Generate a new migration
# migration-new:
# 	@if [ ! -d "db/migrations" ]; then \
# 		mkdir -p db/migrations; \
# 		echo "Created db/migrations directory"; \
# 	fi
# 	@if [ "$(name)" = "" ]; then \
# 		echo "Please provide a migration name using 'make migration-new name=<migration_name>'"; \
# 	else \
# 		migrate create -ext sql -dir db/migrations -seq $(name); \
# 		echo "Migration files created successfully in db/migrations"; \
# 	fi

# # Generate Queries
# sqlc:
# 	sqlc -f db/sqlc.yaml generate