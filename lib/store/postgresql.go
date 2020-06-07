package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"database/sql"

	// Postgres db library loading
	_ "github.com/lib/pq"
)

// PostgresqlStore is a storage engine that writes to postgres
type PostgresqlStore struct {
	db *sql.DB
}

// NewPostgresqlClient creates a new db client object
func NewPostgresqlClient(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		CREATE TABLE IF NOT EXISTS users (
			id STRING NOT NULL,
			username STRING NOT NULL,
			token STRING NOT NULL,
			updated timestamp with time zone NOT NULL,
			PRIMARY KEY(id)
		)
	`)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewPostgresqlStore creates new store
func NewPostgresqlStore(db *sql.DB) PostgresqlStore {
	return PostgresqlStore{
		db: db,
	}
}

// Ping will check if the connection works right
func (s PostgresqlStore) Ping(ctx context.Context) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.PingContext(ctx)
}

// WriteUser will write a user object to postgres
func (s PostgresqlStore) WriteUser(user *User) error {
	if user.ID == "" {
		return fmt.Errorf("user can not be empty")
	}

	_, err := s.db.Exec(`
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
func (s PostgresqlStore) GetUser(id string) (*User, error) {
	var username string
	var tokenJSON string
	var updated time.Time

	err := s.db.QueryRow("SELECT username, token, updated FROM users WHERE id=$1", id).Scan(&username, &tokenJSON, &updated)
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
