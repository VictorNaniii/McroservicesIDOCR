#!/bin/bash

# Test script to send ID images to the OCR service via Kafka

echo "Testing OCR Microservice with ID images..."
echo "=========================================="

# Image files to test
IMAGES=(
  "/home/victornani/Documents/golang/Moldovan_ID_card_(plastic_part_front_side,_1996_year_model).jpg"
  "/home/victornani/Documents/golang/prenume.jpg"
  "/home/victornani/Documents/golang/eca-2020.png"
)

# Copy images to a location accessible by the container
echo "Copying test images to container-accessible location..."
docker cp "/home/victornani/Documents/golang/Moldovan_ID_card_(plastic_part_front_side,_1996_year_model).jpg" awesomeproject-ocr-service-1:/tmp/ocr-images/test-id-1.jpg
docker cp "/home/victornani/Documents/golang/prenume.jpg" awesomeproject-ocr-service-1:/tmp/ocr-images/test-id-2.jpg
docker cp "/home/victornani/Documents/golang/eca-2020.png" awesomeproject-ocr-service-1:/tmp/ocr-images/test-id-3.png

echo ""
echo "Starting consumer in background to capture results..."
docker compose exec -d kafka kafka-console-consumer \
  --bootstrap-server localhost:29092 \
  --topic id-scan-results \
  --from-beginning > /tmp/ocr-results.log 2>&1 &

CONSUMER_PID=$!

sleep 2

echo ""
echo "Sending test requests to Kafka..."
echo ""

# Test 1: Moldovan ID Card
echo "Test 1: Moldovan ID Card"
echo '{"request_id":"test-001","image_data":"","image_path":"/tmp/ocr-images/test-id-1.jpg"}' | \
  docker compose exec -T kafka kafka-console-producer \
  --broker-list localhost:29092 \
  --topic id-scan-requests

sleep 3

# Test 2: Prenume image
echo "Test 2: Prenume ID"
echo '{"request_id":"test-002","image_data":"","image_path":"/tmp/ocr-images/test-id-2.jpg"}' | \
  docker compose exec -T kafka kafka-console-producer \
  --broker-list localhost:29092 \
  --topic id-scan-requests

sleep 3

# Test 3: ECA 2020
echo "Test 3: ECA 2020 Document"
echo '{"request_id":"test-003","image_data":"","image_path":"/tmp/ocr-images/test-id-3.png"}' | \
  docker compose exec -T kafka kafka-console-producer \
  --broker-list localhost:29092 \
  --topic id-scan-requests

echo ""
echo "Waiting for processing to complete..."
sleep 5

echo ""
echo "=========================================="
echo "Checking results from Kafka topic..."
echo "=========================================="
echo ""

# Read results from Kafka
docker compose exec kafka kafka-console-consumer \
  --bootstrap-server localhost:29092 \
  --topic id-scan-results \
  --from-beginning \
  --max-messages 3 \
  --timeout-ms 5000 2>/dev/null

echo ""
echo "=========================================="
echo "Checking OCR service logs..."
echo "=========================================="
echo ""

docker compose logs ocr-service --tail=30

echo ""
echo "Test complete!"
