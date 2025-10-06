#!/bin/bash

# Test script to send sample ID scan requests to Kafka

KAFKA_BROKER="${KAFKA_BROKER:-localhost:9092}"
TOPIC="${TOPIC:-id-scan-requests}"

echo "Sending test message to Kafka..."
echo "Broker: $KAFKA_BROKER"
echo "Topic: $TOPIC"
echo ""

# Sample request with image path
REQUEST=$(cat <<EOF
{
  "request_id": "test-$(date +%s)",
  "image_data": "",
  "image_path": "/path/to/id/card.jpg"
}
EOF
)

echo "Request payload:"
echo "$REQUEST"
echo ""

# Send to Kafka using kafka-console-producer
echo "$REQUEST" | docker-compose exec -T kafka kafka-console-producer \
  --broker-list kafka:29092 \
  --topic "$TOPIC"

echo ""
echo "Message sent! Check the results topic for output."
echo ""
echo "To monitor results, run:"
echo "docker-compose exec kafka kafka-console-consumer --bootstrap-server kafka:29092 --topic id-scan-results --from-beginning"
