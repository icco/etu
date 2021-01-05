package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/icco/etu"
	"github.com/icco/etu/cmd/etu/location"
	"github.com/icco/graphql/time/hexdate"
	"github.com/icco/graphql/time/neralie"
	"github.com/urfave/cli/v2"
)

func (cfg *Config) Add(c *cli.Context) error {
	loc, err := location.CurrentLocation()
	if err != nil {
		log.Printf("could not get location: %+v", err)
	}

	client, err := cfg.Client(c.Context)
	if err != nil {
		return err
	}

	if cfg.slug == "" {
		cfg.slug = fmt.Sprintf("%s/%s", hexdate.Now().String(), neralie.Now().String())
	}

	p, err := etu.GetPage(c.Context, client, cfg.slug)
	if err != nil {
		return err
	}

	p.Meta.Set("latitude", strconv.FormatFloat(loc.Coordinate.Latitude, 'f', -1, 64))
	p.Meta.Set("longitude", strconv.FormatFloat(loc.Coordinate.Longitude, 'f', -1, 64))
	p.Meta.Set("type", "journal")

	tmpl, err := etu.ToMarkdown(p)
	if err != nil {
		return fmt.Errorf("to md: %w", err)
	}

	content, err := CaptureInputFromEditor(tmpl.Bytes())
	if err != nil {
		return fmt.Errorf("get input: %w", err)
	}

	page, err := etu.FromMarkdown(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("from md: %w", err)
	}

	if err := etu.EditPage(c.Context, client, page.Slug, page.Content, page.Meta); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	log.Printf("uploaded n://%s", page.Slug)
	return nil
}
