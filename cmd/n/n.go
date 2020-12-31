// Command n syncs everything from etu.
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/icco/etu"
	"github.com/machinebox/graphql"
	"github.com/urfave/cli/v2"
)

type Config struct {
	APIKey string
	Env    string
	Dir    string
}

func main() {
	cfg := &Config{}
	app := &cli.App{
		Name:  "n",
		Usage: "Wiki sync",
		Commands: []*cli.Command{
			{
				Name:    "sync",
				Aliases: []string{"s"},
				Action:  cfg.Sync,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "api_key",
				Usage:       "authorize your user",
				EnvVars:     []string{"GQL_TOKEN"},
				Destination: &cfg.APIKey,
			},
			&cli.StringFlag{
				Name:        "env",
				Usage:       "set which graphql server to talk to",
				Value:       "production",
				EnvVars:     []string{"NAT_ENV"},
				Destination: &cfg.Env,
			},
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "set where to store the wiki",
				Value:       "~/wiki",
				Destination: &cfg.Dir,
			},
		},
	}

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func (cfg *Config) Path(filename string) string {
	path, _ := filepath.Abs(filepath.Join(cfg.Dir, filename))
	return path
}

func (cfg *Config) Client(ctx context.Context) (*graphql.Client, error) {
	url := ""
	switch cfg.Env {
	case "production":
		url = "https://graphql.natwelch.com/graphql"
	case "development":
		url = "http://localhost:9393/graphql"
	default:
		return nil, fmt.Errorf("unknown environment %q", cfg.Env)
	}

	return etu.NewGraphQLClient(ctx, url, cfg.APIKey)
}

func (cfg *Config) Sync(c *cli.Context) error {
	client, err := cfg.Client(c.Context)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.Path("")); err != nil {
		return err
	}

	pages, err := etu.GetPages(c.Context, client)
	if err != nil {
		return err
	}

	for _, p := range pages {
		path := cfg.Path(p.Slug)
		if err := ioutil.WriteFile(path, p.Content, 0644); err != nil {
			return fmt.Errorf("could not write %q: %w", path, err)
		}
	}
}
