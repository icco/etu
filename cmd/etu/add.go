package main

import (
	"fmt"
	"log"

	"github.com/icco/etu"
	"github.com/icco/etu/cmd/etu/location"
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

	tmpl := fmt.Sprintf("\n\n\nLocation: %+v\n", loc.Coordinate)
	content, err := CaptureInputFromEditor([]byte(tmpl))
	if err != nil {
		return fmt.Errorf("get input: %w", err)
	}

	return etu.EditPage(c.Context, client, cfg.slug, string(content))
}
