# ID OCR Microservice

A high-performance microservice that performs OCR (Optical Character Recognition) on ID documents and extracts structured data such as first name, last name, birth date, and IDNP (Personal Identification Number). Built with Go and integrated with Apache Kafka for asynchronous message processing.

## Features

- **OCR Processing**: Uses Tesseract OCR to extract text from ID document images
- **Data Extraction**: Intelligent parsing of ID information including:
  - First Name
  - Last Name
  - Birth Date
  - IDNP (13-digit Personal Identification Number)
- **Kafka Integration**:
  - Consumes scan requests from Kafka topic
  - Publishes results to Kafka topic
- **Scalable Architecture**: Worker-based processing with configurable concurrency
- **Docker Support**: Fully containerized with Docker Compose
- **Graceful Shutdown**: Handles termination signals properly
- **Comprehensive Logging**: JSON-formatted structured logging
- **Auto Cleanup**: Automatic cleanup of temporary files

## Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   Producer  │────────▶│  Kafka Broker    │────────▶│   Consumer  │
│  (External) │         │  Topic: requests │         │ OCR Service │
└─────────────┘         └──────────────────┘         └──────┬──────┘
                                                             │
                                                             ▼
                                                    ┌────────────────┐
                                                    │  Tesseract OCR │
                                                    │  ID Extraction │
                                                    └────────┬───────┘
                                                             │
                                                             ▼
                        ┌──────────────────┐         ┌──────────────┐
                        │  Kafka Broker    │◀────────│   Producer   │
                        │  Topic: results  │         │  OCR Service │
                        └────────┬─────────┘         └──────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Consumer       │
                        │  (External)     │
                        └─────────────────┘
```

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- Tesseract OCR (if running locally without Docker)

## Installation

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd id-ocr-service
```

2. Start all services:
```bash
docker-compose up -d
```

This will start:
- Zookeeper
- Kafka broker
- Kafka UI (accessible at http://localhost:8090)
- OCR microservice

### Local Development

1. Install dependencies:
```bash
# Install Tesseract OCR
# Ubuntu/Debian
sudo apt-get install tesseract-ocr tesseract-ocr-eng

# macOS
brew install tesseract

# Fedora
sudo dnf install tesseract tesseract-langpack-eng
```

2. Install Go dependencies:
```bash
go mod download
```

3. Update configuration:
Edit `config/config.yaml` to match your environment.

4. Run the service:
```bash
go run cmd/ocr-service/main.go -config config/config.yaml
```

## Configuration

Edit `config/config.yaml`:

```yaml
kafka:
  brokers:
    - "localhost:9092"
  consumer:
    group_id: "id-ocr-consumer-group"
    topic: "id-scan-requests"
  producer:
    topic: "id-scan-results"

ocr:
  tesseract_data_path: "/usr/share/tesseract-ocr/4.00/tessdata"
  language: "eng"
  temp_dir: "/tmp/ocr-images"

service:
  name: "id-ocr-service"
  log_level: "info"  # debug, info, warn, error
  workers: 5
```

## Usage

### Sending Scan Requests

Send messages to the `id-scan-requests` Kafka topic with the following JSON format:

```json
{
  "request_id": "unique-request-id-123",
  "image_data": "<base64-encoded-image>",
  "image_path": ""
}
```

Or use a file path:

```json
{
  "request_id": "unique-request-id-456",
  "image_data": "",
  "image_path": "/path/to/id/image.jpg"
}
```

### Receiving Results

Listen to the `id-scan-results` Kafka topic for responses:

```json
{
  "request_id": "unique-request-id-123",
  "success": true,
  "data": {
    "first_name": "JOHN",
    "last_name": "DOE",
    "birth_date": "01.01.1990",
    "idnp": "1234567890123",
    "raw_text": "Full OCR extracted text...",
    "timestamp": "2024-01-01T12:00:00Z"
  },
  "error": ""
}
```

### Example with kafka-console-producer

```bash
# Produce a test message
echo '{"request_id":"test-001","image_path":"/path/to/id.jpg"}' | \
kafka-console-producer --broker-list localhost:9092 --topic id-scan-requests

# Consume results
kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic id-scan-results --from-beginning
```

## Testing

### Sample Request with Base64 Image

```bash
# Encode image to base64
base64 -w 0 sample-id.jpg > encoded.txt

# Create JSON request
cat << EOF > request.json
{
  "request_id": "test-$(date +%s)",
  "image_data": "$(cat encoded.txt)",
  "image_path": ""
}
EOF

# Send to Kafka
kafka-console-producer --broker-list localhost:9092 \
  --topic id-scan-requests < request.json
```

## Project Structure

```
.
├── cmd/
│   └── ocr-service/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── kafka/
│   │   ├── config.go            # Kafka configuration
│   │   ├── consumer.go          # Kafka consumer implementation
│   │   └── producer.go          # Kafka producer implementation
│   ├── models/
│   │   └── id_data.go           # Data models
│   └── ocr/
│       └── service.go           # OCR processing logic
├── pkg/
│   └── logger/
│       └── logger.go            # Logging utilities
├── config/
│   └── config.yaml              # Service configuration
├── Dockerfile                   # Docker image definition
├── docker-compose.yaml          # Multi-container setup
├── go.mod                       # Go module dependencies
└── README.md                    # This file
```

## Monitoring

### Kafka UI

Access Kafka UI at http://localhost:8090 to:
- View topics and messages
- Monitor consumer lag
- Inspect message contents
- Check broker health

### Logs

View service logs:
```bash
# Docker
docker-compose logs -f ocr-service

# Local
# Logs output to stdout in JSON format
```

## Performance Tuning

1. **Adjust worker count**: Modify `service.workers` in config to match your CPU cores
2. **Kafka consumer group**: Scale horizontally by running multiple instances with the same `consumer.group_id`
3. **OCR optimization**: Pre-process images (resize, enhance contrast) before sending to improve accuracy and speed

## Troubleshooting

### OCR not extracting data correctly

- Ensure image quality is good (minimum 300 DPI recommended)
- Check if the ID document has clear, readable text
- Verify Tesseract language packs are installed
- Try different OCR languages if document is not in English

### Kafka connection issues

- Verify Kafka broker is running: `docker-compose ps`
- Check broker configuration in `config/config.yaml`
- Ensure topics exist or auto-create is enabled

### Service won't start

- Check logs: `docker-compose logs ocr-service`
- Verify Tesseract is installed (for local development)
- Ensure temp directory has write permissions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License

## Support

For issues and questions, please open an issue on GitHub.
