package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/icco/etu"
	"github.com/icco/graphql/time/hexdate"
	"github.com/icco/graphql/time/neralie"
	"github.com/urfave/cli/v2"
)

func (cfg *Config) Add(c *cli.Context) error {
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

	if cfg.file != "" {
		// do upload
		path, err := etu.UploadImage(c.Context, cfg.APIKey, cfg.file)
		if err != nil {
			return err
		}

		raw := path.String()
		path.RawQuery = "auto=format%2Ccompress"

		log.Printf("got path: %v", path)
		p.Content = fmt.Sprintf("[![](%s)](%s)\n\n", path.String(), raw) + p.Content
	}

	if p.Meta.Get("type") == "" {
		p.Meta.Set("type", "journal")
	}

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

	log.Printf("uploaded https://etu.natwelch.com/page/%s", page.Slug)
	return nil
}
