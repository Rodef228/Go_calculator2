package main

import (
	"fmt"
	"http_calculator/internals/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Post("/api/v1/calculate", handlers.CalculateHandler)
	r.Get("/api/v1/expressions", handlers.ExpressionsHandler)
	r.Get("/api/v1/expressions/{id}", handlers.ExpressionHandler)
	r.Get("/internal/task", handlers.TaskHandler)
	r.Post("/internal/task", handlers.TaskResultHandler)

	r.Handle("/*", http.FileServer(http.Dir("./web")))

	fmt.Println("Orchestrator running on :8080")
	http.ListenAndServe(":8080", r)
}
