// Etu is the personifcation of time according to the Lakota.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/icco/etu"
	gql "github.com/icco/graphql"
	"github.com/machinebox/graphql"
	"github.com/urfave/cli/v2"
)

// Config stores all of our settings to run our cmd line app.
type Config struct {
	APIKey string
	Env    string

	project string
}

func main() {
	cfg := &Config{}
	app := &cli.App{
		Name:  "etu",
		Usage: "Log time",
		Commands: []*cli.Command{
			{
				Name:    "timer",
				Usage:   "record time for a project",
				Aliases: []string{"t"},
				Action:  cfg.Timer,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "project",
						Usage:       "project to log for",
						Destination: &cfg.project,
					},
				},
			},
			{
				Name:    "pomodoro",
				Usage:   "record a pomodoro for a project",
				Aliases: []string{"p"},
				Action:  cfg.Pomodoro,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "project",
						Usage:       "project to log for",
						Destination: &cfg.project,
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
		log.Printf("error running: %+v", err)
		os.Exit(1)
	}
}

// Client generates our gql client.
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

func (cfg *Config) Upload(ctx context.Context, start, stop time.Time, sector gql.WorkSector, project, description string) {
	client, err := cfg.Client(ctx)
	if err != nil {
		log.Printf("error creating client: %+v", err)
	}

	if err := etu.UploadLog(ctx, client, &gql.NewLog{
		Sector:      sector,
		Project:     project,
		Description: &description,
		Started:     start,
		Stopped:     stop,
	}); err != nil {
		log.Printf("error uploading: %+v", err)
	}
}
