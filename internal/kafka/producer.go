package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"github.com/victornani/id-ocr-service/internal/models"
)

type Producer struct {
	config   *Config
	producer sarama.SyncProducer
	logger   *logrus.Logger
}

func NewProducer(config *Config, logger *logrus.Logger) (*Producer, error) {
	saramaConfig := NewSaramaConfig()

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &Producer{
		config:   config,
		producer: producer,
		logger:   logger,
	}, nil
}

func (p *Producer) SendResult(response *models.ScanResponse) error {
	messageBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.config.ProducerTopic,
		Key:   sarama.StringEncoder(response.RequestID),
		Value: sarama.ByteEncoder(messageBytes),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Errorf("Failed to send message: %v", err)
		return err
	}

	p.logger.Infof("Message sent successfully: partition=%d, offset=%d", partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
