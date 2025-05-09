package main

import (
	"calculator/internal/agent"
	"calculator/pkg/config"
)

func main() {
	cfg := config.Load()

	agent := agent.New(cfg)
	agent.Run()
}
