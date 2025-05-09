package database

import (
	"calculator/pkg/models"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// DB — обёртка над соединением с базой данных
type DB struct {
	*pgx.Conn
}

// NewDB создаёт новое соединение с PostgreSQL
func NewDB(connString string) (*DB, error) {
	if connString == "" {
		return nil, fmt.Errorf("connection string is empty")
	}

	log.Println("Connecting to PostgreSQL...")

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	// Проверяем соединение
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Conn: conn}, nil
}

// // GetStdlibDB возвращает *sql.DB для работы с миграциями или пулом соединений
// func (db *DB) GetStdlibDB() *sql.DB {
// 	return stdlib.GetSQLDBFromPool(db.Conn.PgConn().Context())
// }

// Close закрывает соединение с базой данных
func (db *DB) Close() {
	if db == nil || db.Conn == nil {
		log.Println("Attempted to close a nil or uninitialized DB connection")
		return
	}
	db.Conn.Close(context.Background())
	log.Println("Database connection closed")
}

// InsertUser добавляет нового пользователя в базу данных
func (db *DB) InsertUser(ctx context.Context, user *models.User) (int, error) {
	if db == nil || db.Conn == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	login := user.Login
	passwordHash := user.Password

	var id int
	err := db.QueryRow(ctx, `
        INSERT INTO users (login, password_hash)
        VALUES ($1, $2)
        RETURNING id`, login, passwordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return id, nil
}

// SelectUserByLogin выбирает пользователя по логину
func (db *DB) SelectUserByLogin(ctx context.Context, login string) (int, string, error) {
	if db == nil || db.Conn == nil {
		return 0, "", fmt.Errorf("database connection is nil")
	}

	var id int
	var passwordHash string

	err := db.QueryRow(ctx, `
        SELECT id, password_hash FROM users WHERE login = $1`, login).Scan(&id, &passwordHash)
	if err != nil {
		return 0, "", fmt.Errorf("failed to select user by login: %w", err)
	}

	return id, passwordHash, nil
}

// UpdateExpression обновляет статус и результат выражения
func (db *DB) UpdateExpression(ctx context.Context, id int, status string, result float64) error {
	if db == nil || db.Conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	_, err := db.Exec(ctx, `
        UPDATE expressions SET status = $1, result = $2 WHERE id = $3`, status, result, id)
	if err != nil {
		return fmt.Errorf("failed to update expression: %w", err)
	}

	return nil
}

// SelectExprByID выбирает выражение по ID и UserID
func (db *DB) SelectExprByID(ctx context.Context, exprID, userID int) (string, string, error) {
	if db == nil || db.Conn == nil {
		return "", "", fmt.Errorf("database connection is nil")
	}

	var expression string
	var status string

	err := db.QueryRow(ctx, `
        SELECT expression, status FROM expressions
        WHERE id = $1 AND user_id = $2`, exprID, userID).Scan(&expression, &status)
	if err != nil {
		return "", "", fmt.Errorf("failed to get expression by ID: %w", err)
	}

	return expression, status, nil
}

// SelectExpressions выбирает все выражения пользователя
func (db *DB) SelectExpressions(ctx context.Context, userID int) ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	rows, err := db.Query(ctx, `
        SELECT id, expression, status, result, created_at, finished_at
        FROM expressions
        WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expressions: %w", err)
	}
	defer rows.Close()

	var expressions []map[string]interface{}

	for rows.Next() {
		var id int
		var expression string
		var status string
		var result sql.NullFloat64
		var createdAt string
		var finishedAt sql.NullString

		if err := rows.Scan(&id, &expression, &status, &result, &createdAt, &finishedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		expr := map[string]interface{}{
			"id":          id,
			"expression":  expression,
			"status":      status,
			"created_at":  createdAt,
			"result":      result.Float64,
			"finished_at": finishedAt.String,
		}

		expressions = append(expressions, expr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return expressions, nil
}

// InsertExpression добавляет новое выражение
func (db *DB) InsertExpression(ctx context.Context, userID int, expression string) (int, error) {
	if db == nil || db.Conn == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	var exprID int
	err := db.QueryRow(ctx, `
        INSERT INTO expressions (user_id, expression, status)
        VALUES ($1, $2, 'pending')
        RETURNING id`, userID, expression).Scan(&exprID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert expression: %w", err)
	}

	return exprID, nil
}
