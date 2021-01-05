package etu

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/gernest/front"
	gql "github.com/icco/graphql"
)

const (
	md = `---
slug: "{{.Slug}}"
modified: "{{.Modified | jstime}}"
records:{{range $index, $pair := .Meta.Records }}
  {{$pair.Key}}: "{{$pair.Record}}"{{end}}
---
{{.Content}}
`
)

func ToMarkdown(p *gql.Page) (*bytes.Buffer, error) {
	tmpl, err := template.New("md").Funcs(template.FuncMap{
		"jstime": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}).Parse(md)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, p); err != nil {
		return nil, err
	}

	return &tpl, nil
}

func FromMarkdown(input io.Reader) (*gql.Page, error) {
	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)
	f, body, err := m.Parse(input)
	if err != nil {
		return nil, err
	}
	p := &gql.Page{
		Content: body,
		Meta:    &gql.PageMetaGrouping{},
	}

	if v, ok := f["modified"].(string); ok {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		p.Modified = t
	}

	if v, ok := f["slug"].(string); ok {
		p.Slug = v
	}

	if r, ok := f["records"].(map[interface{}]interface{}); ok {
		for k, v := range r {
			p.Meta.Set(fmt.Sprintf("%v", k), fmt.Sprintf("%v", v))
		}
	}

	return p, nil
}

func GetLinkedSlugs(p *gql.Page) []string {
	r := blackfriday.NewHTMLRenderer(HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	parser := New(blackfriday.WithRenderer(r), blackfriday.WithExtensions(CommonExtensions))
	ast := parser.Parse(input)

	var ret []string
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Link {
			log.Printf("link %+v", node)
			u, err := url.Parse(string(node.Destination))
			if err != nil {
				log.Errorf("parse: %v", err)
				return blackfriday.Terminate
			}

			if u.Scheme == "n" {
				ret = append(ret, u.Host+u.Path)
			}
		}

		return blackfriday.GoToNext
	})

	return ret
}
