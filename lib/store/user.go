package store

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type store interface {
	WriteUser(user *User) error
}

// User object
type User struct {
	ID       string
	Username string
	Token    *oauth2.Token
	Updated  time.Time
	store    store
}

// NewUser creates a new user object
func NewUser(username string, token *oauth2.Token, storage store) (*User, error) {
	id := uuid.New()
	user := &User{
		ID:       id.String(),
		Username: username,
		Token:    token,
		Updated:  time.Now(),
		store:    storage,
	}

	if err := user.save(); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user object
func (user *User) UpdateUser(token *oauth2.Token) error {
	user.Token = token
	user.Updated = time.Now()

	return user.save()
}

func (user *User) save() error {
	return user.store.WriteUser(user)
}
