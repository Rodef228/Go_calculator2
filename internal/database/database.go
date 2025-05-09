package database

import (
	"context"
	"fmt"

	"calculator/pkg/models"

	"github.com/jackc/pgx/v4"
)

type DB struct {
	conn *pgx.Conn
}

// NewDB создает новое соединение с базой данных
func NewDB(connString string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return &DB{conn: conn}, nil
}

// Close закрывает соединение с базой данных
func (db *DB) Close() {
	db.conn.Close(context.Background())
}

// InsertUser добавляет нового пользователя в базу данных
func (db *DB) InsertUser(ctx context.Context, user *models.User) (int, error) {
	var id int
	err := db.conn.QueryRow(ctx, "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id", user.Login, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SelectUserByLogin выбирает пользователя по логину
func (db *DB) SelectUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := db.conn.QueryRow(ctx, "SELECT id, login, password FROM users WHERE login = $1", login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateExpression обновляет выражение в базе данных
func (db *DB) UpdateExpression(ctx context.Context, id int, status string, result float64) error {
	_, err := db.conn.Exec(ctx, "UPDATE expressions SET status = $1, result = $2 WHERE id = $3", status, result, id)
	return err
}

// SelectExprByID выбирает выражение по ID
func (db *DB) SelectExprByID(ctx context.Context, id int, userID int) (*models.Expression, error) {
	var expr models.Expression
	err := db.conn.QueryRow(ctx, "SELECT id, user_id, expression, status, result FROM expressions WHERE id = $1 AND user_id = $2", id, userID).Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Status, &expr.Result)
	if err != nil {
		return nil, err
	}
	return &expr, nil
}

// SelectExpressions выбирает все выражения для пользователя
func (db *DB) SelectExpressions(ctx context.Context, userID int) ([]models.Expression, error) {
	rows, err := db.conn.Query(ctx, "SELECT id, user_id, expression, status, result FROM expressions WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []models.Expression
	for rows.Next() {
		var expr models.Expression
		if err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Status, &expr.Result); err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)
	}
	return expressions, nil
}
