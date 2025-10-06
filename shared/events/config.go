package events

import (
	"os"
)

// Config RabbitMQ配置
type Config struct {
	URL      string
	Exchange string
}

// NewConfig 创建新的配置
func NewConfig() *Config {
	return &Config{
		URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange: getEnv("RABBITMQ_EXCHANGE", "ride_sharing"),
	}
}

// NewTripExchangeConfig 创建Trip交换器配置
func NewTripExchangeConfig() *Config {
	return &Config{
		URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange: getEnv("TRIP_EXCHANGE", "trip"),
	}
}

// NewPaymentExchangeConfig 创建支付交换器配置
func NewPaymentExchangeConfig() *Config {
	return &Config{
		URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange: getEnv("PAYMENT_EXCHANGE", "payment"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}