package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/icco/etu"
	"github.com/urfave/cli/v2"
)

// Path generates a path for a file to store locally.
func (cfg *Config) Path(filename string) string {
	path, _ := filepath.Abs(filepath.Join(cfg.dir, url.PathEscape(filename)))
	return path
}

// Sync downloads and uploads files based on changes to the content in the server and local host.
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

	// TODO: Compare edited files and upload
	for _, p := range pages {
		path := cfg.Path(p.Slug + ".md")

		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create file: %w ", err)
		}
		defer f.Close()

		bb, err := etu.ToMarkdown(p)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		if _, err := bb.WriteTo(f); err != nil {
			return fmt.Errorf("write: %w", err)
		}

		if err := os.Chtimes(path, time.Now(), p.Modified); err != nil {
			return fmt.Errorf("chtimes: %w", err)
		}
	}

	return nil
}
