# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc g++ musl-dev tesseract-ocr-dev leptonica-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ocr-service ./cmd/ocr-service

# Runtime stage
FROM alpine:latest

# Install Tesseract OCR and dependencies
RUN apk update && apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    ca-certificates \
    tzdata

# Create app user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ocr-service .
COPY --from=builder /app/config ./config

# Create temp directory
RUN mkdir -p /tmp/ocr-images && chown -R appuser:appuser /tmp/ocr-images

# Change to non-root user
USER appuser

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the application
CMD ["./ocr-service", "-config", "config/config.yaml"]
