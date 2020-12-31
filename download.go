package etu

import (
	"context"
	"fmt"

	gql "github.com/icco/graphql"
	"github.com/machinebox/graphql"
)

func GetPage(ctx context.Context, client *graphql.Client, slug string) (*gql.Page, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug required to get page")
	}
	req := graphql.NewRequest(`
query GetPage($slug: ID!) {
	page(slug: $slug) {
    slug
    content
    user {
      id
    }
    meta {
      key
      value
    }
    modified
	}
}`)

	req.Var("slug", slug)

	var p *gql.Page
	if err := client.Run(ctx, req, p); err != nil {
		return nil, err
	}

	return p, nil
}

func GetPages(ctx context.Context, client *graphql.Client) ([]*gql.Page, error) {
	req := graphql.NewRequest(`
query GetPages {
	pages(input: {
    limit: 1000
  }) {
    slug
    content
    user {
      id
    }
    meta {
      key
      record
    }
    modified
	}
}`)

	var p []*gql.Page
	if err := client.Run(ctx, req, p); err != nil {
		return nil, err
	}

	return p, nil
}
