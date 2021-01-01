package etu

import (
	"bytes"
	"html/template"

	gql "github.com/icco/graphql"
)

const (
	md = `---
slug: {{.Slug}}
modified: {{.Modified}}
records:
{{range $index, $pair := .Meta.Records }}
  {{$pair.Key}}: {{$pair.Record}}
{{end}}
---
{{.Content}}
`
)

var (
	tmpl = template.Must(template.New("md").Parse(md))
)

func ToMarkdown(p *gql.Page) (*bytes.Buffer, error) {

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, p); err != nil {
		return nil, err
	}

	return &tpl, nil
}
