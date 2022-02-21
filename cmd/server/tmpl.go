package main

import "html/template"

var (
	indexTmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://unpkg.com/tachyons/css/tachyons.min.css">
    <link rel="webmention" href="https://webmention.io/natwelch.com/webmention" />
    <link rel="pingback" href="https://webmention.io/natwelch.com/xmlrpc" />
  </head>
  <body>
    <article class="pa3 pa5-ns">
      <h1 class="">{{ .Header }}</h1>
      <div class="measure lh-copy">
        {{ .Content }}
      </div>
    </article>
  </body>
</html>`))
)
