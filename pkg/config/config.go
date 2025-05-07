package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort          string
	AddTimeMs           time.Duration
	SubtractTimeMs      time.Duration
	MultiplyTimeMs      time.Duration
	DivideTimeMs        time.Duration
	AgentComputingPower int
}

func Load() *Config {
	return &Config{
		ServerPort:          getEnv("PORT", "8080"),
		AddTimeMs:           time.Duration(getEnvInt("TIME_ADDITION_MS", 1000)) * time.Millisecond,
		SubtractTimeMs:      time.Duration(getEnvInt("TIME_SUBTRACTION_MS", 1000)) * time.Millisecond,
		MultiplyTimeMs:      time.Duration(getEnvInt("TIME_MULTIPLICATIONS_MS", 1000)) * time.Millisecond,
		DivideTimeMs:        time.Duration(getEnvInt("TIME_DIVISIONS_MS", 1000)) * time.Millisecond,
		AgentComputingPower: getEnvInt("COMPUTING_POWER", 10),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
