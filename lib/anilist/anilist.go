package anilist

import (
	"context"
	"fmt"

	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
)

// Handle decides what API calls to make based off of the incoming Plex
// webhook.
func Handle(ctx context.Context, pr plexhooks.PlexResponse, user *store.User) error {
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
