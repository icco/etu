package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/icco/gutil/logging"
)

const (
	// GQLDomain is the target of our requests.
	GQLDomain = "https://graphql.natwelch.com/graphql"

	// GCPProjectID is the project ID where we should send errors.
	GCPProjectID = "icco-cloud"
)

var (
	log = logging.Must(logging.NewLogger("etu"))
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infow("Starting up", "host", fmt.Sprintf("http://localhost:%s", port))

	r := chi.NewRouter()
	r.Use(logging.Middleware(log.Desugar(), GCPProjectID))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ok")
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := &pageData{
			Content: template.HTML(`Etu is a work in progress. <a href="https://github.com/icco/etu">github.com/icco/etu</a> for more information.`),
			Title:   "Etu: icco's time log",
			Header:  "Etu",
		}

		if err := indexTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		workDir, _ := os.Getwd()
		filesDir := http.Dir(filepath.Join(workDir, "cmd/server/public"))
		fs := http.FileServer(filesDir)
		fs.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+port, r))
}
