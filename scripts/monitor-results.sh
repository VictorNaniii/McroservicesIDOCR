#!/bin/bash

# Script to monitor OCR results from Kafka

KAFKA_BROKER="${KAFKA_BROKER:-kafka:29092}"
TOPIC="${TOPIC:-id-scan-results}"

echo "Monitoring Kafka topic: $TOPIC"
echo "Press Ctrl+C to stop"
echo "----------------------------------------"

docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server "$KAFKA_BROKER" \
  --topic "$TOPIC" \
  --from-beginning \
  --property print.key=true \
  --property key.separator=" => "
