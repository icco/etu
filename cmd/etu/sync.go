package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"github.com/icco/etu"
	"github.com/urfave/cli/v2"
)

const (
	md = `---
slug: {{.Slug}}
modified: {{.Modified}}
records:
{{range $index, $pair := .Meta.Records }}
  {{$pair.Key}}: {{$pair.Record}}
{{end}}
---
{{.Content}}
`
)

var (
	tmpl = template.Must(template.New("md").Parse(md))
)

func (cfg *Config) Path(filename string) string {
	path, _ := filepath.Abs(filepath.Join(cfg.dir, url.PathEscape(filename)))
	return path
}

func (cfg *Config) Sync(c *cli.Context) error {
	client, err := cfg.Client(c.Context)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.Path(""), 0777); err != nil {
		return err
	}

	pages, err := etu.GetPages(c.Context, client)
	if err != nil {
		log.Printf("error getting pages: %v", err)
		return err
	}
	log.Printf("got %d pages", len(pages))

	for _, p := range pages {
		path := cfg.Path(p.Slug + ".md")

		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create file: %w ", err)
		}
		defer f.Close()

		log.Printf("writing to %q: %+v", path, p)
		if err := tmpl.Execute(f, p); err != nil {
			return fmt.Errorf("could not write %q: %w", path, err)
		}
	}

	return nil
}
