package main

import (
	"calculator/internal/orchestrator"
)

func main() {
	app := orchestrator.New()
	app.Run()
}
