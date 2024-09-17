package postgresql

import (
	"context"
	"fmt"
	"os"
	"prism-sso/internal/models"
	"prism-sso/internal/storage"
	"time"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	conn *pgx.Conn
}

var timeout = 5 * time.Second

func New() (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URI"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &Storage{conn: conn}, err
}

func (s *Storage) Stop() error {
	return s.conn.Close(context.Background())
}

func (s *Storage) CreateUser(ctx context.Context, email string, passwordHash string) (int64, error) {
	const prefix = "storage.postgresql.CreateUser"

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %s", prefix, err.Error())
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", prefix, err)
	}

	if exists {
		return 0, fmt.Errorf("%s: %w", prefix, storage.ErrUserExists)
	}

	var id int64
	err = tx.QueryRow(ctx, "INSERT INTO users(email, password_hash, role) VALUES($1,$2,$3)").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", prefix, err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, fmt.Errorf("%s: %v", prefix, err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const prefix = "storage.postgresql.GetUser"

	var user models.User
	err := s.conn.QueryRow(ctx, "SELECT id, email, password_hash, role, is_active, created_at FROM users WHERE email=$1", email).Scan(&user)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", prefix, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", prefix, err)
	}

	return user, nil
}
