package anilist

import (
	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
)

func AuthRequest(root, username, code, refreshToken, grantType string) map[string]interface{} {
	return nil
}

func Handle(pr plexhooks.PlexResponse, user *store.User) {
}

func HandleMovie(pr plexhooks.PlexResponse, accessToken string) {
}

func HandleShow(pr plexhooks.PlexResponse, accessToken string) {
}
