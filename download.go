package etu

import (
	"context"
	"fmt"
	"log"

	gql "github.com/icco/graphql"
	"github.com/machinebox/graphql"
)

type getPageResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		Page *gql.Page `json:"page"`
	} `json:"data"`
}

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
      record
    }
    modified
	}
}`)

	req.Var("slug", slug)

	var resp getPageResponse
	if err := client.Run(ctx, req, &resp); err != nil {
		return nil, err
	}
	log.Printf("got response: %+v", resp)

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("query error: %s", resp.Errors[0])
	}

	return resp.Data.Page, nil
}

type getPagesResponse struct {
	Data struct {
		Pages []*gql.Page `json:"pages"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
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

	var resp getPagesResponse
	if err := client.Run(ctx, req, &resp); err != nil {
		return nil, err
	}
	log.Printf("got response: %+v", resp)

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("query error: %s", resp.Errors[0])
	}

	return resp.Data.Pages, nil
}
