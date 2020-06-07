package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// User object
type User struct {
	ID       string
	Username string
	Token    *oauth2.Token
	Updated  time.Time
	store    Store
}

// NewUser creates a new user object
func NewUser(ctx context.Context, username string, token *oauth2.Token, storage Store) (*User, error) {
	id := uuid.New()
	user := &User{
		ID:       id.String(),
		Username: username,
		Token:    token,
		Updated:  time.Now(),
		store:    storage,
	}

	if err := user.store.WriteUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user object
func (user *User) UpdateUser(ctx context.Context, token *oauth2.Token) error {
	user.Token = token
	user.Updated = time.Now()

	return user.store.WriteUser(ctx, user)
}
