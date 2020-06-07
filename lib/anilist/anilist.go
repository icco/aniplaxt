package anilist

import (
	"context"
	"fmt"
	"log"

	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
	"golang.org/x/oauth2"
)

// Handle decides what API calls to make based off of the incoming Plex
// webhook.
func Handle(ctx context.Context, pr plexhooks.PlexResponse, user *store.User, ts oauth2.TokenSource) error {
	log.Printf("recieved %+v", pr)
	switch pr.Metadata.LibrarySectionType {
	case "show":
		return HandleShow(ctx, pr, user, ts)
	case "movie":
		return HandleMovie(ctx, pr, user, ts)
	}

	return fmt.Errorf("unknown type")
}

// HandleMovie handles a plex movie watch.
func HandleMovie(ctx context.Context, pr plexhooks.PlexResponse, user *store.User, ts oauth2.TokenSource) error {
	return fmt.Errorf("unimplemented")
}

// HandleShow handles a plex tv episode watch.
func HandleShow(ctx context.Context, pr plexhooks.PlexResponse, user *store.User, ts oauth2.TokenSource) error {
	return fmt.Errorf("unimplemented")
}
