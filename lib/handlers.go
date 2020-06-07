package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/icco/aniplaxt/lib/anilist"
	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
)

// AuthorizePage is a data struct for authorized pages.
type AuthorizePage struct {
	SelfRoot   string
	Authorized bool
	URL        string
	ClientID   string
}

// SelfRoot gets the root url we are serving from.
func SelfRoot(r *http.Request) string {
	u := &url.URL{}
	u.Host = r.Host
	u.Scheme = r.URL.Scheme
	if u.Scheme == "" {
		u.Scheme = "http"

		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "https" {
			u.Scheme = "https"
		}
	}
	return u.String()
}

// Authorize is a handler for users to log in and store their authorization information.
func Authorize(storage store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Query()
		username := strings.ToLower(args["username"][0])
		log.Print(fmt.Sprintf("Handling auth request for %s", username))
		code := args["code"][0]
		result, err := anilist.AuthRequest(SelfRoot(r), username, code, "", "authorization_code")
		if err != nil {
			log.Errorf("could not auth: %+v", err)
			http.Error(w, "something went wrong with auth", http.StatusInternalServerError)
			return
		}

		user := store.NewUser(username, result["access_token"].(string), result["refresh_token"].(string), storage)

		url := fmt.Sprintf("%s/api?id=%s", SelfRoot(r), user.ID)

		log.Print(fmt.Sprintf("Authorized as %s", user.ID))

		tmpl := template.Must(template.ParseFiles("static/index.html"))
		data := AuthorizePage{
			SelfRoot:   SelfRoot(r),
			Authorized: true,
			URL:        url,
			ClientID:   os.Getenv("ANILIST_ID"),
		}
		tmpl.Execute(w, data)
	}
}

// API is the handler which parses Plex webhook requests.
func API(storage store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Query()
		id := args["id"][0]
		log.Debugf("webhook call for %q", id)

		user := storage.GetUser(id)

		tokenAge := time.Since(user.Updated).Hours()
		if tokenAge > 1440 { // tokens expire after 3 months, so we refresh after 2
			log.Debugf("User access token outdated, refreshing...")
			result, err := anilist.AuthRequest(SelfRoot(r), user.Username, "", user.RefreshToken, "refresh_token")
			if err != nil {
				log.Errorf("could not auth: %+v", err)
				http.Error(w, "something went wrong with auth", http.StatusInternalServerError)
				return
			}

			user.UpdateUser(result["access_token"].(string), result["refresh_token"].(string))
			log.Debugf("Refreshed, continuing")
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("could not read body: %+v", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		regex := regexp.MustCompile("({.*})") // not the best way really
		match := regex.FindStringSubmatch(string(body))
		re, err := plexhooks.ParseWebhook([]byte(match[0]))
		if err != nil {
			log.Errorf("could not parse body: %+v", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		if strings.ToLower(re.Account.Title) == user.Username {
			log.Errorf("Plex username %s does not equal %s, skipping", strings.ToLower(re.Account.Title), user.Username)
			json.NewEncoder(w).Encode("wrong user")
			return
		}

		if err := anilist.Handle(re, user); err != nil {
			log.Errorf("could not handle: %+v", err)
			http.Error(w, "something went wrong talking to anilist", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("success")
	}
}

// AllowedHostsHandler is a middleware which takes a comma seperated lists of
// hostnames and filters requests so those without a Host header with a value
// in the list recieve a 403. /healthz is whitelisted.
func AllowedHostsHandler(allowedHostnames string) func(http.Handler) http.Handler {
	allowedHosts := strings.Split(regexp.MustCompile("https://|http://|\\s+").ReplaceAllString(strings.ToLower(allowedHostnames), ""), ",")
	log.Infof("Allowed Hostnames: %v", allowedHosts)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.EscapedPath() == "/healthz" {
				h.ServeHTTP(w, r)
				return
			}
			isAllowedHost := false
			lcHost := strings.ToLower(r.Host)
			for _, value := range allowedHosts {
				if lcHost == value {
					isAllowedHost = true
					break
				}
			}
			if !isAllowedHost {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprintf(w, "Oh no!")
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
