package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
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

var DB database.DB

func logsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %s, URL: %s", r.Method, r.URL)
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("Method: %s, completion time: %v", r.Method, duration)
	})
}

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
				log.Printf("Code: %v, user unauthorized", http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				errorResponse(w, "invalid token format", http.StatusUnauthorized)
				log.Printf("Code: %v, invalid token format", http.StatusUnauthorized)
				return
			}
			token = tokenParts[1]
			log.Print("token was taken from header")
		}

		claims, id := jwt.Verify(token)
		if !claims {
			errorResponse(w, "invalid token", http.StatusUnauthorized)
			log.Printf("Code: %v, invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func databaseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "DB", DB)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, "invalid request method", http.StatusMethodNotAllowed)
		log.Printf("Code: %v, invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest)
		log.Printf("Code: %v, json decoding error", http.StatusBadRequest)
		return
	}

	if len(body.Password) == 0 {
		errorResponse(w, "password cannot be empty", http.StatusForbidden)
		log.Printf("Code: %v, empty password", http.StatusForbidden)
		return
	}

	pass, err := password.Generate(body.Password)
	if err != nil {
		errorResponse(w, "internal server error", http.StatusInternalServerError)
		log.Printf("Code: %v, %s", http.StatusInternalServerError, err)
		return
	}

	ctx := r.Context()
	user := &models.User{
		Login:    body.Login,
		Password: pass,
	}
	_, err = DB.InsertUser(ctx, user) // Используем DB для вставки пользователя
	if err != nil {
		errorResponse(w, "user already exists", http.StatusConflict)
		log.Printf("Code: %v, user %s already exists", http.StatusConflict, body.Login)
		return
	}

	log.Printf("user: %v has successfully registered", user.Login)
	w.WriteHeader(http.StatusOK)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, "invalid request method", http.StatusMethodNotAllowed)
		log.Printf("Code: %v, invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest)
		log.Printf("Code: %v, json decoding error", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := DB.SelectUserByLogin(ctx, body.Login) // Используем DB для выбора пользователя
	if err != nil {
		errorResponse(w, "user not found", http.StatusNotFound)
		log.Printf("Code: %v, user %v was not found", http.StatusNotFound, body.Login)
		return
	}
	if err := password.Compare(user.Password, body.Password); err != nil {
		errorResponse(w, "incorrect password", http.StatusForbidden)
		log.Printf("Code: %v, incorrect password", http.StatusForbidden)
		return
	}

	var resp struct {
		Jwt string `json:"jwt"`
	}
	token, err := jwt.Generate(int(user.ID))
	if err != nil {
		errorResponse(w, "internal server error", http.StatusInternalServerError)
		log.Printf("Code: %v, error with generating token", http.StatusInternalServerError)
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

	resp.Jwt = token
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ExpressionHandler(w http.ResponseWriter, r *http.Request) {
	expr := r.Context().Value(ctxKey).(*Expression)

	// создаем аст по выражению
	astRoot, err := ast.Build(expr.exp)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)

		log.Printf("Expression %d: AST build failed - %s", expr.id, errStr)
		if err := DB.UpdateExpression(context.Background(), expr.id, errStr, 0.0); err != nil {
			log.Printf("Failed to update expression %d: %v", expr.id, err)
		}
		return
	}
	exp := NewExpression(astRoot)

	res, err := exp.calc()
	if err != nil {
		log.Printf("Expression %v: zero division error detected", expr.id)
		if err := DB.UpdateExpression(context.Background(), expr.id, "zero division error", 0.0); err != nil {
			log.Printf("Failed to update expression %d: %v", expr.id, err)
		}
		return
	}
	log.Printf("Expression %v calculated successfully", expr.id)

	if err := DB.UpdateExpression(context.Background(), expr.id, "done", res); err != nil {
		log.Printf("Failed to update expression %d: %v", expr.id, err)
	}
}

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	if checkId(id) {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errStr := fmt.Sprintf("%s", err)
			errorResponse(w, "internal server error", http.StatusInternalServerError)
			log.Printf("Code: %v, %s", http.StatusInternalServerError, errStr)
			return
		}

		userId := r.Context().Value(userID)
		data, err := DB.SelectExprByID(r.Context(), idInt, userId.(int)) // Используем db для выбора выражения по ID
		if err != nil {
			errorResponse(w, "expression does not exist", http.StatusNotFound)
			log.Printf("Code: %v, %s", http.StatusNotFound, err)
			return
		}

		// Сериализация данных в JSON
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			errorResponse(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
			log.Printf("Code: %v, error with marshaling json", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData) // Отправляем сериализованные данные пользователю
		if err != nil {
			errorResponse(w, "error with json data", http.StatusInternalServerError)
			log.Printf("Code: %v, Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	data, err := DB.SelectExpressions(r.Context(), r.Context().Value(userID).(int)) // Используем db для выбора всех выражений пользователя
	if err != nil {
		errorResponse(w, "you haven't calculated any expressions yet", http.StatusInternalServerError)
		log.Printf("Code: %v, empty base for user %v", http.StatusInternalServerError, userID)
		return
	}

	// Сериализация данных в JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		errorResponse(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		log.Printf("Code: %v, error with marshaling json", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // устанавливаем статус 200 OK
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData) // Отправляем сериализованные данные пользователю
	if err != nil {
		errorResponse(w, "error with json data", http.StatusInternalServerError)
		log.Printf("Code: %v, Internal server error", http.StatusInternalServerError)
		return
	}
}
