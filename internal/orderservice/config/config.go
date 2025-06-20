package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort   int
	DatabaseURL  string
	KafkaBrokers []string
}

func LoadConfig() (*Config, error) {
	//Server Port
	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		portStr = "8080" // Default port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	//Database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is not set")
	}

	//Kafka Brokers
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		return nil, errors.New("KAFKA_BROKERS environment variable is not set")
	}
	kafkaBrokers := splitAndTrim(kafkaBrokersStr, ",")

	return &Config{
		ServerPort:   port,
		DatabaseURL:  dbURL,
		KafkaBrokers: kafkaBrokers,
	}, nil
}

func splitAndTrim(s, sep string) []string {
	var result []string
	parts := strings.Split(s, sep)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
