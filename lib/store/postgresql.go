package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

// PostgresqlStore is a storage engine that writes to postgres
type PostgresqlStore struct {
	db *pgx.ConnConfig
}

// NewPostgresqlStore creates new store
func NewPostgresqlStore(connStr string) (*PostgresqlStore, error) {
	ctx := context.Background()
	cfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	db, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT NOT NULL,
			username TEXT NOT NULL,
			token TEXT NOT NULL,
			updated timestamp with time zone NOT NULL,
			PRIMARY KEY(id)
		)
	`)
	if err != nil {
		return nil, err
	}
	return &PostgresqlStore{
		db: cfg,
	}, nil
}

// Ping will check if the connection works right
func (s *PostgresqlStore) Ping(ctx context.Context) error {
	conn, err := pgx.ConnectConfig(ctx, s.db)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	return conn.Ping(ctx)
}

// WriteUser will write a user object to postgres
func (s *PostgresqlStore) WriteUser(ctx context.Context, user *User) error {
	if user.ID == "" {
		return fmt.Errorf("user can not be empty")
	}

	conn, err := pgx.ConnectConfig(ctx, s.db)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `
			INSERT INTO users (id, username, token, updated) VALUES ($1, $2, $3, $4)
			ON CONFLICT(id)
			DO UPDATE set username=EXCLUDED.username, token=EXCLUDED.token, updated=EXCLUDED.updated`,
		user.ID,
		user.Username,
		user.Token,
		user.Updated,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetUser will load a user from postgres
func (s *PostgresqlStore) GetUser(ctx context.Context, id string) (*User, error) {
	conn, err := pgx.ConnectConfig(ctx, s.db)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	var username string
	var tokenJSON []byte
	var updated time.Time

	err = conn.QueryRow(ctx, "SELECT username, token, updated FROM users WHERE id=$1", id).Scan(&username, &tokenJSON, &updated)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("no user with id %s", id)
	case err != nil:
		return nil, fmt.Errorf("query error: %w", err)
	}
	user := User{
		ID:       id,
		Username: strings.ToLower(username),
		Token:    JSONToToken(tokenJSON),
		Updated:  updated,
		store:    s,
	}

	return &user, nil
}
