package etu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

func UploadImage(ctx context.Context, apikey, path string) (*url.URL, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	filetype := mime.TypeByExtension(filepath.Ext(path))
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filetype, filepath.Base(file.Name()))
	if err != nil {
		return nil, err
	}

	io.Copy(part, file)
	writer.Close()

	request, err := http.NewRequest("POST", "https://graphql.natwelch.com/photo/new", body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization", fmt.Sprint("Bearer %s", apikey))
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("sure: %+v", content)

	return nil, fmt.Errorf("unimplemented")
}

func EditPage(ctx context.Context, client *graphql.Client, slug, content string, meta *gql.PageMetaGrouping) error {

	gql := `
mutation SavePage($content: String!, $slug: ID!, $meta: [InputMeta]!) {
  upsertPage(input: {
    content: $content,
    slug: $slug,
    meta: $meta,
  }) {
    modified
  }
}`

	req := graphql.NewRequest(gql)
	req.Var("content", content)
	req.Var("slug", slug)
	req.Var("meta", meta.Records)

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("edit page %+v: %w", req, err)
	}

	return nil
}
