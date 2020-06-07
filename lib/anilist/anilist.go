package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
)

// AuthRequest parses the auth request to anilist.
func AuthRequest(root, username, code, refreshToken, grantType string) (map[string]string, error) {
	values := map[string]string{
		"code":          code,
		"refresh_token": refreshToken,
		"client_id":     os.Getenv("ANILIST_ID"),
		"client_secret": os.Getenv("ANILIST_SECRET"),
		//"redirect_uri":  fmt.Sprintf("%s/authorize?username=%s", root, url.PathEscape(username)),
		"redirect_uri": fmt.Sprintf("%s/authorize", root),
		"grant_type":   grantType,
	}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 200 {
		//log.WithField("response", resp).Warnf("got a %q. Aborting to avoid panic.", resp.Status)
		return nil, fmt.Errorf("could not contact anilist API: %+v", resp)
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
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
