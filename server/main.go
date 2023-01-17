package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/icco/gutil/logging"
	"github.com/unrolled/render"
)

var (
	log          = logging.Must(logging.NewLogger("etu"))
	GCPProjectID = "icco-cloud"

	// Renderer is a renderer for all occasions. These are our preferred default options.
	// See:
	//  - https://github.com/unrolled/render/blob/v1/README.md
	//  - https://godoc.org/gopkg.in/unrolled/render.v1
	Renderer = render.New(render.Options{
		Charset:                   "UTF-8",
		Directory:                 "./server/views",
		DisableHTTPErrorRendering: true,
		Extensions:                []string{".tmpl", ".html"},
		Funcs:                     []template.FuncMap{{}},
		IndentJSON:                false,
		IndentXML:                 true,
		Layout:                    "layout",
		RequirePartials:           true,
	})
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))
	r.Use(logging.Middleware(log.Desugar(), GCPProjectID))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/icco/etu", 301)
	})
	r.Get("/healthz", healthCheckHandler)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	Renderer.JSON(w, http.StatusOK, map[string]string{
		"healthy": "true",
	})
}
