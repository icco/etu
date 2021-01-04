package main

import (
	"fmt"
	"html/template"
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

type pageData struct {
	Content   template.HTML
	Title     string
	Header    string
	SubHeader string
	Page      *gql.Page
}

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

		data := &pageData{
			Content: template.HTML(blackfriday.Run([]byte(resp.Page.Content))),
			Title:   fmt.Sprintf("Etu: %q", resp.Page.Slug),
			Header:  resp.Page.Slug,
			Page:    resp.Page,
		}

		if err := pageTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Fatalln(http.ListenAndServe(":"+port, r))
}
