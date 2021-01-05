// Etu is the personifcation of time according to the Lakota.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/icco/etu"
	"github.com/machinebox/graphql"
	"github.com/urfave/cli/v2"
)

type Config struct {
	APIKey string
	Env    string
	slug   string
	dir    string
}

func main() {
	cfg := &Config{}
	app := &cli.App{
		Name:  "etu",
		Usage: "Journaling from the command line",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a log",
				Action:  cfg.Add,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "slug",
						Aliases:     []string{"s"},
						Usage:       "slug to save page as",
						Destination: &cfg.slug,
					},
				},
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "generate missing slugs",
				Action:  cfg.Generate,
			},
			{
				Name:    "sync",
				Aliases: []string{"s"},
				Usage:   "Sync wiki to disk",
				Action:  cfg.Sync,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "dir",
						Usage:       "set where to store the wiki",
						Value:       "/tmp/wiki",
						Destination: &cfg.dir,
					},
				},
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
		},
	}

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
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
