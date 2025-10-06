package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Kafka   KafkaConfig   `yaml:"kafka"`
	OCR     OCRConfig     `yaml:"ocr"`
	Service ServiceConfig `yaml:"service"`
}

type KafkaConfig struct {
	Brokers  []string       `yaml:"brokers"`
	Consumer ConsumerConfig `yaml:"consumer"`
	Producer ProducerConfig `yaml:"producer"`
}

type ConsumerConfig struct {
	GroupID         string `yaml:"group_id"`
	Topic           string `yaml:"topic"`
	AutoOffsetReset string `yaml:"auto_offset_reset"`
}

type ProducerConfig struct {
	Topic string `yaml:"topic"`
}

type OCRConfig struct {
	TesseractDataPath string `yaml:"tesseract_data_path"`
	Language          string `yaml:"language"`
	TempDir           string `yaml:"temp_dir"`
}

type ServiceConfig struct {
	Name     string `yaml:"name"`
	LogLevel string `yaml:"log_level"`
	Workers  int    `yaml:"workers"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Service.Workers == 0 {
		config.Service.Workers = 5
	}

	if config.OCR.Language == "" {
		config.OCR.Language = "eng"
	}

	if config.OCR.TempDir == "" {
		config.OCR.TempDir = "/tmp/ocr-images"
	}

	return &config, nil
}
