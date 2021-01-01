package etu

import (
	"context"
	"fmt"
	"net/http"

	gql "github.com/icco/graphql"
	"github.com/machinebox/graphql"
)

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

func NewGraphQLClient(ctx context.Context, endpoint, apikey string) (*graphql.Client, error) {
	httpclient := &http.Client{
		Transport: &AddHeaderTransport{T: http.DefaultTransport, Key: apikey},
	}
	client := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpclient))

	gql := `
  query {
    whoami {
			id
    }
  }`
	req := graphql.NewRequest(gql)

	var response struct {
		WhoAmI struct {
			ID string
		}
	}

	if err := client.Run(ctx, req, &response); err != nil {
		return nil, err
	}

	if response.WhoAmI.ID == "" {
		return nil, fmt.Errorf("invalid user")
	}

	return client, nil
}

func EditPage(ctx context.Context, client *graphql.Client, slug, content string, meta *gql.PageMetaGrouping) error {

	gql := `
mutation SavePage($content: String!, $slug: ID!) {
	upsertPage(input: {
    content: $content,
    slug: $slug,
	}) {
    modified
	}
}`

	req := graphql.NewRequest(gql)
	req.Var("content", content)
	req.Var("slug", slug)

	return client.Run(ctx, req, nil)
}
