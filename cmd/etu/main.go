package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/icco/etu"
	"github.com/icco/etu/cmd/etu/location"
	"github.com/machinebox/graphql"
	"github.com/urfave/cli/v2"
)

type Config struct {
	APIKey string
	Env    string
	slug   string
}

// Etu is the personifcation of time according to the Lakota.
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

type AddHeaderTransport struct {
	T   http.RoundTripper
	Key string
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if adt.Key == "" {
		return nil, fmt.Errorf("no key provided")
	}

	req.Header.Add("X-API-AUTH", adt.Key)

	return adt.T.RoundTrip(req)
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

	return etu.EditPage(ctx, client, cfg.slug, string(content))
}
