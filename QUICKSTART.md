# Quick Start Guide

Get the ID OCR microservice up and running in 5 minutes!

## Step 1: Start the Services

```bash
# Start all services (Kafka, Zookeeper, OCR Service)
docker-compose up -d

# Wait for services to be ready (about 30 seconds)
docker-compose ps
```

## Step 2: Verify Services are Running

```bash
# Check service logs
docker-compose logs -f ocr-service

# You should see: "Service started successfully"
```

## Step 3: Access Kafka UI (Optional)

Open your browser and navigate to:
```
http://localhost:8090
```

You can view topics, messages, and monitor the service.

## Step 4: Send a Test Request

### Option A: Using the test script

```bash
./scripts/test-producer.sh
```

### Option B: Manual Kafka console

```bash
# Send a test message
echo '{
  "request_id": "test-001",
  "image_data": "",
  "image_path": "/path/to/your/id.jpg"
}' | docker-compose exec -T kafka kafka-console-producer \
  --broker-list kafka:29092 \
  --topic id-scan-requests
```

### Option C: Using image data (base64 encoded)

```bash
# Encode your ID image
base64 -w 0 your-id.jpg > encoded.txt

# Create and send request
cat << EOF | docker-compose exec -T kafka kafka-console-producer --broker-list kafka:29092 --topic id-scan-requests
{
  "request_id": "test-$(date +%s)",
  "image_data": "$(cat encoded.txt)",
  "image_path": ""
}
EOF
```

## Step 5: Monitor Results

### Option A: Using the monitor script

```bash
./scripts/monitor-results.sh
```

### Option B: Manual Kafka console

```bash
docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server kafka:29092 \
  --topic id-scan-results \
  --from-beginning
```

Expected output:
```json
{
  "request_id": "test-001",
  "success": true,
  "data": {
    "first_name": "JOHN",
    "last_name": "DOE",
    "birth_date": "01.01.1990",
    "idnp": "1234567890123",
    "raw_text": "Full OCR text...",
    "timestamp": "2024-01-01T12:00:00Z"
  },
  "error": ""
}
```

## Step 6: Scale the Service (Optional)

```bash
# Scale to 3 instances
docker-compose up -d --scale ocr-service=3

# Verify
docker-compose ps
```

## Common Commands

```bash
# View logs
docker-compose logs -f ocr-service

# Restart service
docker-compose restart ocr-service

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Rebuild after code changes
docker-compose up -d --build
```

## Troubleshooting

### Service won't start

```bash
# Check logs
docker-compose logs ocr-service

# Check if Kafka is ready
docker-compose exec kafka kafka-broker-api-versions --bootstrap-server kafka:29092
```

### No results appearing

1. Check if request was received:
```bash
docker-compose logs ocr-service | grep "Processing scan request"
```

2. Check consumer group status:
```bash
docker-compose exec kafka kafka-consumer-groups \
  --bootstrap-server kafka:29092 \
  --describe \
  --group id-ocr-consumer-group
```

3. Verify topics exist:
```bash
docker-compose exec kafka kafka-topics \
  --bootstrap-server kafka:29092 \
  --list
```

## Next Steps

- Check out the [README.md](README.md) for detailed documentation
- Integrate with your application using Kafka clients
- Customize OCR patterns in `internal/ocr/service.go`
- Adjust configuration in `config/config.yaml`

## Production Checklist

- [ ] Update Kafka brokers to production cluster
- [ ] Configure proper authentication (SASL/SSL)
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Test with production-like ID images
- [ ] Tune worker count based on load
- [ ] Set up horizontal pod autoscaling (if using K8s)
- [ ] Configure resource limits
- [ ] Set up backup for Kafka topics
- [ ] Implement health check endpoints
