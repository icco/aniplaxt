package store

import (
	"context"
	"encoding/json"
	"log"

	"golang.org/x/oauth2"
)

// Store is the interface for All the store types
type Store interface {
	WriteUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	Ping(ctx context.Context) error
}

// TokenToJSON serializes an oauth2 token.
func TokenToJSON(t *oauth2.Token) []byte {
	tokJSON, err := json.Marshal(t)
	if err != nil {
		log.Printf("could not marshal %+v", t)
		panic(err)
	}

	return tokJSON
}

// JSONToToken turns stored JSON into a token.
func JSONToToken(s []byte) *oauth2.Token {
	var tok oauth2.Token
	if err := json.Unmarshal([]byte(s), &tok); err != nil {
		log.Printf("could not unmarshal %q", s)
		panic(err)
	}
	return &tok
}
