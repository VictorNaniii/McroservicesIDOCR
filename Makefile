.PHONY: build run test clean docker-build docker-up docker-down deps

# Build the application
build:
	go build -o bin/ocr-service ./cmd/ocr-service

# Run the application locally
run:
	go run cmd/ocr-service/main.go -config config/config.yaml

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf /tmp/ocr-images/*

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker build -t id-ocr-service:latest .

# Start Docker Compose services
docker-up:
	docker-compose up -d

# Stop Docker Compose services
docker-down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f ocr-service

# Rebuild and restart service
restart: docker-down docker-build docker-up

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Create Kafka topics manually
create-topics:
	docker-compose exec kafka kafka-topics --create --topic id-scan-requests --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1
	docker-compose exec kafka kafka-topics --create --topic id-scan-results --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1

# List Kafka topics
list-topics:
	docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:29092

# Monitor consumer group
monitor-consumer:
	docker-compose exec kafka kafka-consumer-groups --bootstrap-server localhost:29092 --describe --group id-ocr-consumer-group
