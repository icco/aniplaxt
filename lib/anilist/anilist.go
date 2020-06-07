package anilist

import (
	"fmt"

	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
)

func AuthRequest(root, username, code, refreshToken, grantType string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func Handle(pr plexhooks.PlexResponse, user *store.User) error {
	return fmt.Errorf("unimplemented")
}

func HandleMovie(pr plexhooks.PlexResponse, accessToken string) error {
	return fmt.Errorf("unimplemented")
}

func HandleShow(pr plexhooks.PlexResponse, accessToken string) error {
	return fmt.Errorf("unimplemented")
}
