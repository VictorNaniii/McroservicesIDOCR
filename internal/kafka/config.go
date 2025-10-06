package kafka

import (
	"github.com/IBM/sarama"
)

type Config struct {
	Brokers       []string
	ConsumerGroup string
	ConsumerTopic string
	ProducerTopic string
}

// NewSaramaConfig creates a new Sarama configuration
func NewSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V3_3_0_0

	// Consumer settings
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	// Producer settings
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	return config
}
