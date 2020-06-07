package anilist

import (
	"fmt"

	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
)

// AuthRequest parses the auth request to anilist.
func AuthRequest(root, username, code, refreshToken, grantType string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

// Handle decides what API calls to make based off of the incoming Plex
// webhook.
func Handle(pr plexhooks.PlexResponse, user *store.User) error {
	return fmt.Errorf("unimplemented")
}

// HandleMovie handles a plex movie watch.
func HandleMovie(pr plexhooks.PlexResponse, accessToken string) error {
	return fmt.Errorf("unimplemented")
}

// HandleShow handles a plex tv episode watch.
func HandleShow(pr plexhooks.PlexResponse, accessToken string) error {
	return fmt.Errorf("unimplemented")
}
