package main

import (
	"calculator/internal/application"
	"calculator/pkg/config"
	"calculator/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New("orchestrator")

	app := application.New(cfg, log)
	app.Run()
}
