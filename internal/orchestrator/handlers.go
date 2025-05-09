package orchestrator

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"calculator/internal/database"
	"calculator/pkg/ast"
	"calculator/pkg/models"
	"calculator/pkg/pass_system/jwt"
	"calculator/pkg/pass_system/password"
)

// Middleware для логирования
func logsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %s, URL: %s", r.Method, r.URL)
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("Method: %s, completion time: %v", r.Method, duration)
	})
}

// Middleware для аутентификации
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		cookie, err := r.Cookie("jwt")
		if checkCookie(cookie, err) {
			token = cookie.Value
			log.Print("token was taken from cookie")
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorResponse(w, "authorization is required", http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				errorResponse(w, "invalid token format", http.StatusUnauthorized)
				return
			}
			token = tokenParts[1]
		}

		claims, id := jwt.Verify(token)
		if !claims {
			errorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware для передачи DB в контекст
func databaseMiddleware(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "DB", db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Регистрация пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request, db *database.DB) {
	if r.Method != http.MethodPost {
		errorResponse(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(body.Password) == 0 {
		errorResponse(w, "password cannot be empty", http.StatusForbidden)
		return
	}

	pass, err := password.Generate(body.Password)
	if err != nil {
		errorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Login:    body.Login,
		Password: pass,
	}

	id, err := db.InsertUser(r.Context(), user)
	if err != nil {
		errorResponse(w, "user already exists", http.StatusConflict)
		return
	}

	log.Printf("user: %v has successfully registered (ID: %d)", user.Login, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

// Авторизация пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request, db *database.DB) {
	if r.Method != http.MethodPost {
		errorResponse(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, pass, err := db.SelectUserByLogin(r.Context(), body.Login)
	if err != nil {
		errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	if err := password.Compare(pass, body.Password); err != nil {
		errorResponse(w, "incorrect password", http.StatusForbidden)
		return
	}

	token, err := jwt.Generate(int(id))
	if err != nil {
		errorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(10 * time.Minute),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jwt": token})
}

// Вычисление выражения
func ExpressionHandler(w http.ResponseWriter, r *http.Request, db *database.DB) {
	var req struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Создаём выражение локально
	expr := &Expression{
		exp: req.Expression,
		id:  0, // можно генерировать при сохранении в БД
	}

	astRoot, err := ast.Build(expr.exp)
	if err != nil {
		db.UpdateExpression(r.Context(), expr.id, err.Error(), 0)
		errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	calc := NewExpression(astRoot)
	result, err := calc.calc()
	if err != nil {
		db.UpdateExpression(r.Context(), expr.id, "zero division error", 0)
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Сохраняем результат в БД (если нужно)
	db.UpdateExpression(r.Context(), expr.id, "done", result)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]float64{"result": result})
}

// Получение данных по ID или всех выражений
func GetDataHandler(w http.ResponseWriter, r *http.Request, db *database.DB) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	if checkId(path) {
		idInt, _ := strconv.Atoi(path)
		userId := r.Context().Value(userID).(int)

		expr, _, err := db.SelectExprByID(r.Context(), idInt, userId)
		if err != nil {
			errorResponse(w, "expression does not exist", http.StatusNotFound)
			return
		}

		jsonData, _ := json.MarshalIndent(expr, "", "  ")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	userId := r.Context().Value(userID).(int)
	data, err := db.SelectExpressions(r.Context(), userId)
	if err != nil {
		errorResponse(w, "you haven't calculated any expressions yet", http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
