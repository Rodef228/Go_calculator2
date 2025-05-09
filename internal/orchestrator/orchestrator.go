package orchestrator

import (
	"calculator/internal/database"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"

	"github.com/go-chi/chi/v5"
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
	} else {
		log.Printf("Connected to database")
	}
	defer o.db.Close(context.Background())

	if err := runMigrations(); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	// запуск менеджера каналов выражений
	StartManager()
	// запуск сервера для общения с агентом
	go runGRPC()

	db, err := database.NewDB(DB_URL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	r := chi.NewRouter()
	r.Use(logsMiddleware)

	// Передаём db в хендлеры
	r.Post("/api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, db)
	})
	r.Post("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, db)
	})

	calculateRouter := chi.NewRouter()
	calculateRouter.Use(authMiddleware)
	calculateRouter.Use(databaseMiddleware(db)) // передаём db в middleware
	calculateRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
		expr := &Expression{
			exp: "2+2",
			id:  123,
		}
		ctx := context.WithValue(r.Context(), ctxKey, expr)
		ExpressionHandler(w, r.WithContext(ctx), db)
	})

	r.Mount("/api/v1/calculate", calculateRouter)

	r.With(authMiddleware).Get("/api/v1/expressions/", func(w http.ResponseWriter, r *http.Request) {
		GetDataHandler(w, r, db)
	})

	log.Printf("Starting server on port '%s'", orchURL)
	log.Fatal(http.ListenAndServe(orchURL, r))
}
