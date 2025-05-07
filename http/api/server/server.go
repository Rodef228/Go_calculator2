package server

import (
	"context"
	"net/http"
	"time"

	"calculator/http/api/handler"
	"calculator/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux
	server *http.Server
	log    logger.Logger
}

func New(h *handler.TaskHandler, log logger.Logger) *Server {
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

	return &Server{
		router: r,
		log:    log,
	}
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	s.log.Infow("Server started on %s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
