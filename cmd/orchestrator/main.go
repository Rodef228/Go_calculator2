package main

import (
	"calculator/internal/handler"
	"calculator/internal/service"
	"calculator/pkg/config"
	"calculator/pkg/logger"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang-migrate/migrate"
)

type Application struct {
	config config.Config
	logi   logger.Logger
}

type Server struct {
	router *chi.Mux
	server *http.Server
	logg   logger.Logger
}

func runMigrations() error {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	m, err := migrate.New(
		"file://migrations",
		dbURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	s.logg.Infow("Server started on %s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func main() {
	cfg := config.Load()
	loog := logger.New("orchestrator")

	app := &Application{
		config: cfg,
		logi:   loog,
	}

	taskService := service.NewTaskService()
	h := handler.NewTaskHandler(taskService)

	if err := runMigrations(); err != nil {
		log.Fatal("Migrations failed: ", err)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Get("/", h.Index)
	r.Route("/internal", func(r chi.Router) {
		r.Post("/task", h.CreateTask)
		r.Get("/task", h.GetTask)
		r.Put("/task/{id}", h.UpdateTask)
	})

	srv := &Server{
		router: r,
		logg:   app.logi,
	}

	go func() {
		if err := srv.Start(":" + app.config.ServerPort); err != nil && err != http.ErrServerClosed {
			app.logi.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.logi.Error("server shutdown error", "error", err)
	}

	app.logi.Info("server stopped")
}
