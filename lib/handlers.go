package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/icco/aniplaxt"
	"github.com/icco/aniplaxt/lib/anilist"
	"github.com/icco/aniplaxt/lib/store"
	"github.com/icco/gutil/logging"
	"github.com/xanderstrike/plexhooks"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var (
	log = logging.Must(logging.NewLogger(aniplaxt.Service))
)

// AuthorizePage is a data struct for authorized pages.
type AuthorizePage struct {
	Authorized bool
	User       bool
	URL        string
	AuthURL    string
	Token      string
}

// EmptyPageData is a generator for a simple page data that is empty.
func EmptyPageData(r *http.Request) *AuthorizePage {
	url := AuthData(SelfRoot(r)).AuthCodeURL("state", oauth2.AccessTypeOffline)
	return &AuthorizePage{
		AuthURL: url,
		URL:     "https://aniplaxt.natwelch.com/api?id=generate-your-own-silly",
	}
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

// AuthData generates the config for oauth2 for anilist.
func AuthData(root string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("ANILIST_ID"),
		ClientSecret: os.Getenv("ANILIST_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://anilist.co/api/v2/oauth/authorize",
			TokenURL: "https://anilist.co/api/v2/oauth/token",
		},
		RedirectURL: fmt.Sprintf("%s/authorize", root),
	}
}

// Authorize is a handler for users to log in and store their authorization information.
func Authorize(storage store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Query()
		code := args.Get("code")
		ctx := r.Context()

		conf := AuthData(SelfRoot(r))
		tok, err := conf.Exchange(ctx, code)
		if err != nil {
			log.Errorw("could not exchange code", zap.Error(err))
			http.Error(w, "something went wrong with auth", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("static/index.html")
		if err != nil {
			log.Errorw("could not parse index", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		data := EmptyPageData(r)
		data.Authorized = true
		data.Token = base64.StdEncoding.EncodeToString(store.TokenToJSON(tok))
		if err := tmpl.Execute(w, data); err != nil {
			log.Errorw("couldn't render template", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}
}

// RegisterUser is a handler for saving users.
func RegisterUser(storage store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.PostFormValue("username")
		tokBase64String := r.PostFormValue("token")
		ctx := r.Context()

		tokString, err := base64.StdEncoding.DecodeString(tokBase64String)
		if err != nil {
			log.Errorw("could not decode token", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		tok := store.JSONToToken(tokString)

		u, err := store.NewUser(ctx, user, tok, storage)
		if err != nil {
			log.Errorw("could not create user", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("static/index.html")
		if err != nil {
			log.Errorf("could not parse index: %+v", err, zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		data := EmptyPageData(r)
		data.Authorized = true
		data.Token = tok.AccessToken
		data.User = true
		data.URL = fmt.Sprintf("%s/api?id=%s", SelfRoot(r), u.ID)
		if err := tmpl.Execute(w, data); err != nil {
			log.Errorw("couldn't render template", zap.Error(err))
			http.Error(w, "something went wrong with auth", http.StatusInternalServerError)
			return
		}
	}
}

// API is the handler which parses Plex webhook requests.
func API(storage store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args := r.URL.Query()
		id := args.Get("id")
		ctx := r.Context()
		log.Debugf("webhook call for %q", id)

		conf := AuthData(SelfRoot(r))
		user, err := storage.GetUser(ctx, id)
		if err != nil {
			log.Errorw("could not get user", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		tokSource := conf.TokenSource(ctx, user.Token)

		tok, err := tokSource.Token()
		if err != nil {
			log.Errorw("could not get token", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		if err := user.UpdateUser(ctx, tok); err != nil {
			log.Errorw("could not update user", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorw("could not read body", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		// TODO: Figure out wtf this is.
		regex := regexp.MustCompile("({.*})") // not the best way really
		match := regex.FindStringSubmatch(string(body))
		re, err := plexhooks.ParseWebhook([]byte(match[0]))
		if err != nil {
			log.Errorw("could not parse body", zap.Error(err))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		if strings.ToLower(re.Account.Title) != strings.ToLower(user.Username) {
			log.Errorw(fmt.Sprintf("Plex username %q does not equal %q, skipping", strings.ToLower(re.Account.Title), strings.ToLower(user.Username)), "user", user, "account", re.Account)
			json.NewEncoder(w).Encode("wrong user")
			return
		}

		if err := anilist.Handle(ctx, re, user, tokSource); err != nil {
			log.Errorw("could not handle", zap.Error(err))
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
