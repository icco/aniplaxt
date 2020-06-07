package store

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
)

// Store is the interface for All the store types
type Store interface {
	WriteUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	Ping(ctx context.Context) error
}

// TokenToJSON serializes an oauth2 token.
func TokenToJSON(t *oauth2.Token) string {
	tokJson, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}

	return string(tokJson)
}

// JSONToToken turns stored JSON into a token.
func JSONToToken(s string) *oauth2.Token {
	var tok *oauth2.Token
	if err := json.Unmarshal([]byte(s), tok); err != nil {
		panic(err)
	}
	return tok
}
