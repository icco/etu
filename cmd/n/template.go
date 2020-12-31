package main

import "text/template"

const (
	md = `
---
slug: {{.Slug}}
modified: {{.Modified}}
{{range $index, $pair := .Meta }}
{{$pair.Key}}: {{$pair.Record}}
{{end}}
---
{{.Content}}
`
	tmpl = template.Must(template.New("md").Parse(md))
)
