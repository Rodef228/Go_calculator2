package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort          string
	AddTimeMs           int
	SubtractTimeMs      int
	MultiplyTimeMs      int
	DivideTimeMs        int
	AgentComputingPower int
	OrchestratorURL     string
}

func Load() Config {
	return Config{
		ServerPort:          getEnv("PORT", "8080"),
		AddTimeMs:           int(getEnvInt("TIME_ADDITION_MS", 1000)),
		SubtractTimeMs:      int(getEnvInt("TIME_SUBTRACTION_MS", 1000)),
		MultiplyTimeMs:      int(getEnvInt("TIME_MULTIPLICATIONS_MS", 1000)),
		DivideTimeMs:        int(getEnvInt("TIME_DIVISIONS_MS", 1000)),
		AgentComputingPower: getEnvInt("COMPUTING_POWER", 10),
		OrchestratorURL:     getEnv("ORCHESTRATOR_URL", "http://orchestrator:8080"),
	}
}

var Configuration Config

func LoadConfig() {
	Configuration := Load()
	if Configuration.OrchestratorURL == "" {
		Configuration.OrchestratorURL = "http://orchestrator:8080"
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
