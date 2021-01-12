package etu

import (
	"context"
	"fmt"

	gql "github.com/icco/graphql"
	"github.com/machinebox/graphql"
)

type getPageResponse struct {
	Page *gql.Page `json:"page"`
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
      records {
        key
        record
      }
    }
    modified
    created
	}
}`)

	req.Var("slug", slug)

	var resp getPageResponse
	if err := client.Run(ctx, req, &resp); err != nil {
		return nil, err
	}

	if resp.Page.Meta == nil {
		resp.Page.Meta = &gql.PageMetaGrouping{
			Records: []*gql.PageMeta{},
		}
	}

	return resp.Page, nil
}

type getPagesResponse struct {
	Pages []*gql.Page `json:"pages"`
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
    modified
    meta {
      records {
        key
        record
      }
    }
	}
}`)

	var resp getPagesResponse
	if err := client.Run(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Pages, nil
}
