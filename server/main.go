package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/aniplaxt/lib"
	"github.com/icco/aniplaxt/lib/store"
)

var (
	storage store.Store
	log     = logging.Must(logging.NewLogger(aniplaxt.Service))
)

func main() {
	for _, e := range []string{
		"DATABASE_URL",
		"ANILIST_ID",
		"ANILIST_SECRET",
	} {
		if os.Getenv(e) == "" {
			log.Fatalf("%q can not be unset", e)
		}
	}

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infow("Starting up", "host", fmt.Sprintf("http://localhost:%s", port))

	// Connect to db
	dbURL := os.Getenv("DATABASE_URL")
	storage, err := store.NewPostgresqlStore(dbURL)
	if err != nil {
		log.WithError(err).Fatalf("could not connect to %q", dbURL)
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(logging.Middleware(log.Desugar(), *project))

	// which hostnames we are allowing
	// ALLOWED_HOSTNAMES = new accurate config variable
	// No env = all hostnames
	if os.Getenv("ALLOWED_HOSTNAMES") != "" {
		r.Use(lib.AllowedHostsHandler(os.Getenv("ALLOWED_HOSTNAMES")))
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi."))
	})

	r.Get("/authorize", lib.Authorize(storage))
	r.Post("/api", lib.API(storage))
	r.Post("/save", lib.RegisterUser(storage))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("static/index.html")
		if err != nil {
			log.WithError(err).Errorf("could not render template")
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, lib.EmptyPageData(r)); err != nil {
			log.WithError(err).Error("couldn't render template")
		}
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
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
