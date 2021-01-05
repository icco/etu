package main

import (
	"fmt"
	"log"

	"github.com/icco/etu"
	gql "github.com/icco/graphql"
	"github.com/urfave/cli/v2"
)

// Generate looks for all n:// links and makes sure they exist.
func (cfg *Config) Generate(c *cli.Context) error {
	client, err := cfg.Client(c.Context)
	if err != nil {
		return err
	}

	pages, err := etu.GetPages(c.Context, client)
	if err != nil {
		log.Printf("error getting pages: %v", err)
		return err
	}

	var links []string
	slugs := map[string]*gql.Page{}
	for _, p := range pages {
		slugs[p.Slug] = p
		links = append(links, etu.GetLinkedSlugs(p)...)
	}

	for _, l := range links {
		if slugs[l] == nil {
			if err := etu.EditPage(c.Context, client, l, "TBD", &gql.PageMetaGrouping{Records: []*gql.PageMeta{&gql.PageMeta{Key: "type", Record: "stub"}}}); err != nil {
				return fmt.Errorf("uploading %q: %w", l, err)
			}
		}
	}

	return nil
}
