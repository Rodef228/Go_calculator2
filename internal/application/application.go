package application

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"calculator/http/api/handler"
	"calculator/http/api/server"
	"calculator/internal/service"
	"calculator/pkg/config"
	"calculator/pkg/logger"
)

type Application struct {
	config *config.Config
	log    logger.Logger
}

func New(cfg *config.Config, log logger.Logger) *Application {
	return &Application{
		config: cfg,
		log:    log,
	}
}

func (a *Application) Run() {
	// Initialize services
	taskService := service.NewTaskService()
	taskHandler := handler.NewTaskHandler(taskService)

	// Create and start server
	srv := server.New(taskHandler, a.log)

	go func() {
		if err := srv.Start(":" + a.config.ServerPort); err != nil && err != http.ErrServerClosed {
			a.log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.log.Error("server shutdown error", "error", err)
	}

	a.log.Info("server stopped")
}
