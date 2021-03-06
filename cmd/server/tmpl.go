package main

import "html/template"

var (
	pageTmpl = template.Must(template.New("page").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://unpkg.com/tachyons/css/tachyons.min.css">
    <link rel="webmention" href="https://webmention.io/natwelch.com/webmention" />
    <link rel="pingback" href="https://webmention.io/natwelch.com/xmlrpc" />
    <style>
      pre {
        border-left: 3px solid #ddd;
        page-break-inside: avoid;
        font-family: monospace;
        font-size: 15px;
        line-height: 1.6;
        margin-bottom: 1.6em;
        max-width: 100%;
        overflow: auto;
        padding: 1em 1.5em;
        display: block;
        word-wrap: break-word;
      }
    </style>
  </head>
  <body>
    <article class="cf ph3 ph5-ns pv5">
      <header class="fn fl-ns w-50-ns pr4-ns">
        <h1 class="f2 lh-title fw9 mb3 mt0 pt3 bt bw2">
          {{ .Page.Slug }}
        </h1>

        <div class="f6 gray tracked">Created: <time class="ttu">{{ .Page.Created }}</time></div>
        <div class="f6 gray tracked">Modified: <time class="ttu">{{ .Page.Modified }}</time></div>

        <div class="cf">
        {{ range .Page.Meta.Records }}
          <dl class="fn dib w-auto lh-title mr5-l">
            <dd class="f6 fw4 ml0">{{ .Key }}</dd>
            <dd class="fw6 ml0">{{ .Record }}</dd>
          </dl>
        {{ end }}
        </div>

        <div class="cf pb3 bb bw2">
          <h2 class="f6 fw4 ml0 lh-title mt0 pt3">Links to here</h2>
          <ul>
          {{ range .References }}
            <li class="f6 lh-title"><a href="/page/{{ . }}">{{ . }}</a></li>
          {{ else }}
            <li>None</li>
          {{ end }}
          </ul>
        </div>
      </header>
      <div class="fn fl-ns w-50-ns">
        <div class="measure lh-copy">
          {{ .Content }}
        </div>
      </div>
    </article>
  </body>
</html>`))
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
	pagesTmpl = template.Must(template.New("pages").Parse(`
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
        {{ range $key, $value := .Pages }}
            <h2>{{ $key }}</h2>
            <ul class="list pl0 measure center">
              {{ range $value }}
                <li class="lh-copy pv3 ba bl-0 bt-0 br-0 b--dotted b--black-30">
                  <a href="/page/{{ .Slug }}">{{ .Slug }}</a>
                </li>
              {{ end }}
            </ul>
        {{ end }}
      </div>
    </article>
  </body>
</html>`))
)
