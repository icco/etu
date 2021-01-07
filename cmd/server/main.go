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
}

const (
	GQLDomain = "https://graphql.natwelch.com/graphql"
)

var (
	pageTmpl = template.Must(template.New("layout").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://unpkg.com/tachyons/css/tachyons.min.css">
  </head>
  <body>
    <article class="cf ph3 ph5-ns pv5">
      <header class="fn fl-ns w-50-ns pr4-ns">
        <h1 class="f2 lh-title fw9 mb3 mt0 pt3 bt bw2">
          {{ .Header }}
        </h1>
        <h2 class="f3 mid-gray lh-title">{{ .SubHeader }}</h2>
        <time class="f6 ttu tracked gray">{{ .Page.Modified }}</time>
        <div class="cf">
        {{ range .Page.Meta.Records }}
          <dl class="fn dib w-auto lh-title mr5-l">
            <dd class="f6 fw4 ml0">{{ .Key }}</dd>
            <dd class="fw6 ml0">{{ .Record }}</dd>
          </dl>
        {{ end }}
        </div>
      </header>
      <div class="fn fl-ns w-50-ns">
        <div class="measure lh-copy">
          {{ .Content }}
        </div>
      </div>
    </article>
  </body>
</html>`))
	indexTmpl = template.Must(template.New("layout").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://unpkg.com/tachyons/css/tachyons.min.css">
  </head>
  <body>
    <article class="pa3 pa5-ns">
      <h1 class="">{{ .Header }}</h1>
      <div class="measure lh-copy">
        {{ .Content }}
      </div>
    </article>
  </body>
</html>`))
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

		// This is hacky.
		content := `<div class="pa3 pa5-ns"><ul class="list pl0 measure center">`
		for _, p := range pages {
			content += fmt.Sprintf(`<li class="lh-copy pv3 ba bl-0 bt-0 br-0 b--dotted b--black-30"><a href="https://etu.natwelch.com/page/%s">%s</a></li>`, p.Slug, p.Slug)
		}
		content += "</ul></div>"

		data := &pageData{
			Content: template.HTML(content),
			Title:   "Etu: index",
			Header:  "Etu: index",
		}

		if err := indexTmpl.Execute(w, data); err != nil {
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
