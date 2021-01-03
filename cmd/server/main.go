package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gql "github.com/icco/graphql"
	"github.com/machinebox/graphql"
	"github.com/russross/blackfriday/v2"
)

type pageResponse struct {
	Page *gql.Page `json:"page"`
}

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on http://localhost:%s", port)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ok")
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Etu is a work in progress. github.com/icco/etu")
	})

	r.Get("/page/*", func(w http.ResponseWriter, r *http.Request) {
		rawslug := chi.URLParam(r, "*")
		slug, err := url.QueryUnescape(rawslug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query := `query ($slug: ID!) {
      page(slug: $slug) {
        slug
        content
        modified
        meta {
          records {
            key
            record
          }
        }
      }
    }`

		gqlClient := graphql.NewClient("https://graphql.natwelch.com/graphql")
		req := graphql.NewRequest(query)
		req.Header.Add("X-API-AUTH", os.Getenv("GQL_TOKEN"))
		req.Header.Add("User-Agent", "etu-server/1.0")
		req.Var("slug", slug)
		var resp pageResponse
		if err := gqlClient.Run(r.Context(), req, &resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(blackfriday.Run([]byte(resp.Page.Content)))
	})

	log.Fatalln(http.ListenAndServe(":"+port, r))
}
