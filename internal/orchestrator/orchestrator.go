package orchestrator

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4" // Обратите внимание на использование pgx/v4
)

type (
	Orchestrator struct {
		db *pgx.Conn
	}

	ExpressionReq struct {
		Expression string `json:"expression"`
	}

	RespID struct {
		Id int `json:"id"`
	}

	Error struct {
		Res string `json:"error"`
	}

	Expression struct {
		exp string
		id  int
	}

	contextKey string
	userid     string
)

func New() *Orchestrator {
	return &Orchestrator{}
}

var (
	mu     sync.Mutex // Мьютекс для синхронизации доступа к результатам
	ctxKey contextKey = "expression id"
	userID userid     = "user id"
)

func checkCookie(cookie *http.Cookie, err error) bool {
	if err != nil {
		return false
	}

	token := cookie.Value
	return !(len(token) == 0)
}

func errorResponse(w http.ResponseWriter, err string, statusCode int) {
	w.WriteHeader(statusCode)
	e := Error{Res: err}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func checkId(id string) bool {
	pattern := "^[0-9]+$"
	r := regexp.MustCompile(pattern)
	return r.MatchString(id)
}

func (o *Orchestrator) Run() {
	DB_URL := os.Getenv("DB_URL")
	orchURL := os.Getenv("ORCHESTRATOR_URL")

	// Подключение к базе данных
	var err error
	o.db, err = pgx.Connect(context.Background(), DB_URL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer o.db.Close(context.Background())

	// запуск менеджера каналов выражений
	StartManager()
	// запуск сервера для общения с агентом
	go runGRPC()

	// Создание маршрутизатора chi
	r := chi.NewRouter()
	r.Use(middleware.Logger) // Логирование запросов

	// Применяем logsMiddleware ко всем маршрутам
	r.Use(logsMiddleware)

	// Определяем маршруты без дополнительных middleware
	r.Post("/api/v1/register", RegisterHandler)
	r.Post("/api/v1/login", LoginHandler)

	// Создаем новый маршрутизатор для маршрута с несколькими middleware
	calculateRouter := chi.NewRouter()
	calculateRouter.Use(authMiddleware)
	calculateRouter.Use(databaseMiddleware)
	calculateRouter.Post("/", ExpressionHandler)

	// Добавляем новый маршрутизатор для /api/v1/calculate
	r.Mount("/api/v1/calculate", calculateRouter)

	// Определяем маршрут для получения данных с authMiddleware
	r.With(authMiddleware).Get("/api/v1/expressions/", GetDataHandler)

	log.Printf("Starting server on port '%s'", orchURL)
	log.Fatal(http.ListenAndServe(orchURL, r))
}
