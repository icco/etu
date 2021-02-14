package etu

import (
	"bytes"
	"fmt"
	h "html/template"
	"io"
	"log"
	"net/url"
	"text/template"
	"time"

	wikilink "github.com/dangoor/goldmark-wikilinks"
	"github.com/gernest/front"
	gql "github.com/icco/graphql"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
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

type wikilinksExt struct {
	found map[string]string
}

func (wl *wikilinksExt) LinkWithContext(destText string, destFilename string, context string) {
	wl.found[destFilename] = destText
}

func (wl *wikilinksExt) Normalize(in string) string {
	return fmt.Sprintf("/page/%s", url.PathEscape(in))
}

func (wl *wikilinksExt) Extend(m goldmark.Markdown) {
	wlp := wikilink.NewWikilinksParser().WithNormalizer(wl).WithTracker(wl)
	m.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(wlp, 102)),
	)
}

func buildMDParser() (goldmark.Markdown, *wikilinksExt) {
	wl := &wikilinksExt{found: map[string]string{}}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			wl,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	), wl
}

func ToHTML(p *gql.Page) h.HTML {
	var buf bytes.Buffer
	md, _ := buildMDParser()
	if err := md.Convert([]byte(p.Content), &buf); err != nil {
		log.Panic(err)
	}

	return h.HTML(buf.Bytes())
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

func GetLinkedSlugs(p *gql.Page) map[string]bool {
	md, t := buildMDParser()
	var buf bytes.Buffer
	if err := md.Convert([]byte(p.Content), &buf); err != nil {
		log.Panic(err)
	}

	ret := map[string]bool{}
	for k := range t.found {
		ret[k] = true
	}

	return ret
}
