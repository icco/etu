package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/icco/etu"
	gql "github.com/icco/graphql"
)

type pageResponse struct {
	Page *gql.Page `json:"page"`
}

type pageData struct {
	Content   template.HTML
	Title     string
	Header    string
	SubHeader string
	Page      *gql.Page
	Pages     map[string][]*gql.Page
}

const (
	GQLDomain = "https://graphql.natwelch.com/graphql"
)

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
		data := &pageData{
			Content: template.HTML(`Etu is a work in progress. <a href="https://github.com/icco/etu">github.com/icco/etu</a> for more information.`),
			Title:   "Etu: icco's wiki",
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

	r.Get("/pages", func(w http.ResponseWriter, r *http.Request) {
		client, err := etu.NewGraphQLClient(r.Context(), GQLDomain, os.Getenv("GQL_TOKEN"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pages, err := etu.GetPages(r.Context(), client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := &pageData{
			Title:  "Etu: index",
			Header: "Etu: index",
			Pages:  map[string][]*gql.Page{},
		}

		for _, p := range pages {
			t := "unknown"
			if v := p.Meta.Get("type"); v != "" {
				t = v
			}
			data.Pages[t] = append(data.Pages[t], p)
		}

		if err := pagesTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/page/*", func(w http.ResponseWriter, r *http.Request) {
		rawslug := chi.URLParam(r, "*")
		slug, err := url.QueryUnescape(rawslug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client, err := etu.NewGraphQLClient(r.Context(), GQLDomain, os.Getenv("GQL_TOKEN"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		page, err := etu.GetPage(r.Context(), client, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := &pageData{
			Content: etu.ToHTML(page),
			Title:   fmt.Sprintf("Etu: %q", page.Slug),
			Header:  page.Slug,
			Page:    page,
		}

		if err := pageTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Fatalln(http.ListenAndServe(":"+port, r))
}
