package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/icco/aniplaxt/lib"
	"github.com/icco/aniplaxt/lib/anilist"
	"github.com/icco/aniplaxt/lib/store"
	"github.com/xanderstrike/plexhooks"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

var (
	storage store.Store
	log     = lib.InitLogging()
)

type AuthorizePage struct {
	SelfRoot   string
	Authorized bool
	URL        string
	ClientID   string
}

func SelfRoot(r *http.Request) string {
	u, _ := url.Parse("")
	u.Host = r.Host
	u.Scheme = r.URL.Scheme
	u.Path = ""
	if u.Scheme == "" {
		u.Scheme = "http"

		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "https" {
			u.Scheme = "https"
		}
	}
	return u.String()
}

func authorize(w http.ResponseWriter, r *http.Request) {
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

func api(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	id := args["id"][0]
	log.Print(fmt.Sprintf("Webhook call for %s", id))

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

func allowedHostsHandler(allowedHostnames string) func(http.Handler) http.Handler {
	allowedHosts := strings.Split(regexp.MustCompile("https://|http://|\\s+").ReplaceAllString(strings.ToLower(allowedHostnames), ""), ",")
	log.Infof("Allowed Hostnames: %v", allowedHosts)
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.URL.EscapedPath() == "/healthcheck" {
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
		}

		return http.HandlerFunc(fn)
	}
}

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infof("Starting up on http://localhost:%s", port)

	if os.Getenv("ENABLE_STACKDRIVER") != "" {
		labels := &stackdriver.Labels{}
		labels.Set("app", "aniplaxt", "The name of the current app.")
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:               "icco-cloud",
			MonitoredResource:       monitoredresource.Autodetect(),
			DefaultMonitoringLabels: labels,
			DefaultTraceAttributes:  map[string]interface{}{"app": "aniplaxt"},
		})

		if err != nil {
			log.WithError(err).Fatalf("failed to create the stackdriver exporter")
		}
		defer sd.Flush()

		view.RegisterExporter(sd)
		trace.RegisterExporter(sd)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		})
	}

	if os.Getenv("POSTGRESQL_URL") != "" {
		storage = store.NewPostgresqlStore(store.NewPostgresqlClient(os.Getenv("POSTGRESQL_URL")))
		log.Infof("Using postgresql storage: %q", os.Getenv("POSTGRESQL_URL"))
	} else if os.Getenv("REDIS_URI") != "" {
		storage = store.NewRedisStore(store.NewRedisClient(os.Getenv("REDIS_URI"), os.Getenv("REDIS_PASSWORD")))
		log.Infof("Using redis storage: %q", os.Getenv("REDIS_URI"))
	} else {
		storage = store.NewDiskStore()
		log.Infof("Using disk storage")
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(lib.LoggingMiddleware())
	// which hostnames we are allowing
	// ALLOWED_HOSTNAMES = new accurate config variable
	// No env = all hostnames
	if os.Getenv("ALLOWED_HOSTNAMES") != "" {
		r.Use(allowedHostsHandler(os.Getenv("ALLOWED_HOSTNAMES")))
	}
	r.Get("/authorize", authorize)
	r.Post("/api", api)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi."))
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("static/index.html"))
		data := AuthorizePage{
			SelfRoot:   SelfRoot(r),
			Authorized: false,
			URL:        "https://plaxt.astandke.com/api?id=generate-your-own-silly",
			ClientID:   os.Getenv("ANILIST_ID"),
		}
		tmpl.Execute(w, data)
	})

	h := &ochttp.Handler{
		Handler:     r,
		Propagation: &propagation.HTTPFormat{},
	}
	if err := view.Register([]*view.View{
		ochttp.ServerRequestCountView,
		ochttp.ServerResponseCountByStatusCode,
	}...); err != nil {
		log.WithError(err).Fatal("Failed to register ochttp views")
	}

	log.Fatal(http.ListenAndServe(":"+port, h))
}
