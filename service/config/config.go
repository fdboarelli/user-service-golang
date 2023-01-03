package config

import (
	"context"
	"os"
	"strconv"
)

//Config is the Service's configuration object
type Config struct {
	GrpcPort    string
	ServiceName string
	DebugMode   bool
	Mode        string
	MongoHost   string
	SecretKey   string
	KafkaServer string
	KafkaTopic  string
}

// New returns a new Config struct populated with .env values or default ones
func New(ctx context.Context) *Config {
	return &Config{
		GrpcPort:    getEnv("GRPC_PORT", "9090"),
		ServiceName: getEnv("SERVICE_NAME", "user-service"),
		DebugMode:   getEnvAsBool("DEBUG", false),
		Mode:        getEnv("MODE", "DEV"),
		MongoHost:   getEnv("MONGODB_HOST", "mongodb://localhost:27017"),
		SecretKey:   getEnv("SECRET_KEY", "MySecretKey"),
		KafkaServer: getEnv("KAFKA_SERVER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "users_topic"),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}
