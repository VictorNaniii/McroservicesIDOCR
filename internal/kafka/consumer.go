package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"github.com/victornani/id-ocr-service/internal/models"
)

type Consumer struct {
	config        *Config
	consumerGroup sarama.ConsumerGroup
	logger        *logrus.Logger
	handler       MessageHandler
}

type MessageHandler func(ctx context.Context, request *models.ScanRequest) (*models.ScanResponse, error)

func NewConsumer(config *Config, logger *logrus.Logger, handler MessageHandler) (*Consumer, error) {
	saramaConfig := NewSaramaConfig()

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		config:        config,
		consumerGroup: consumerGroup,
		logger:        logger,
		handler:       handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	consumerHandler := &consumerGroupHandler{
		consumer: c,
		ready:    make(chan bool),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := c.consumerGroup.Consume(ctx, []string{c.config.ConsumerTopic}, consumerHandler); err != nil {
				c.logger.Errorf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
			consumerHandler.ready = make(chan bool)
		}
	}()

	<-consumerHandler.ready
	c.logger.Info("Kafka consumer started and ready")

	<-ctx.Done()
	c.logger.Info("Terminating consumer...")

	wg.Wait()
	return c.consumerGroup.Close()
}

type consumerGroupHandler struct {
	consumer *Consumer
	ready    chan bool
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			h.consumer.logger.Infof("Received message: offset=%d, partition=%d", message.Offset, message.Partition)

			var request models.ScanRequest
			if err := json.Unmarshal(message.Value, &request); err != nil {
				h.consumer.logger.Errorf("Failed to unmarshal message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			// Process the message
			response, err := h.consumer.handler(session.Context(), &request)
			if err != nil {
				h.consumer.logger.Errorf("Failed to process message: %v", err)
			}

			// Log the response (producer will handle sending)
			if response != nil {
				responseJSON, _ := json.Marshal(response)
				h.consumer.logger.Debugf("Processing result: %s", string(responseJSON))
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
