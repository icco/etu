# Etu

[![Go Report Card](https://goreportcard.com/badge/github.com/icco/etu)](https://goreportcard.com/report/github.com/icco/etu)
[![Go Reference](https://pkg.go.dev/badge/github.com/icco/etu.svg)](https://pkg.go.dev/github.com/icco/etu)


Etu is a simple journaling tool that talks to the [etu-backend](https://github.com/icco/etu-backend) API over gRPC (default: `grpc.etu.natwelch.com`).

It should be noted the main goal of Etu is to write interstitial journals. See https://betterhumans.pub/replace-your-to-do-list-with-interstitial-journaling-to-increase-productivity-4e43109d15ef for more on this topic.

## Installation

### Homebrew

```shell
brew tap icco/tap
brew install etu
```

### Local Build

**Dependencies:** Go 1.25 or later, [Task](https://taskfile.dev/) (optional).

1. `git clone` the repo
2. `task build` (or `go build -o etu .`)
3. Run `./etu`

## Usage

Before running you need an API key for the etu-backend. You can:

1. Put your API key in `~/.config/etu/config.json` with keys `api_key` and optionally `grpc_target` (default: `grpc.etu.natwelch.com:443`), or
2. Set the `ETU_API_KEY` environment variable (and optionally `ETU_GRPC_TARGET`).

Example config file:

```json
{
  "api_key": "your-64-char-hex-api-key",
  "grpc_target": "grpc.etu.natwelch.com:443"
}
```

Config and the "time since last post" cache live under `~/.config/etu/`. Tag generation and storage are handled by the backend; see [etu-backend](https://github.com/icco/etu-backend) for setup.

```
$ etu
Etu. A personal command line journal.

Usage:
  etu [flags]
  etu [command]

Available Commands:
  create      Create a new journal entry (attach images/audio via drag & drop in TUI or -i/--image, -a/--audio).
  delete      Delete a journal entry.
  help        Help about any command
  last        Output a string of time since last post.
  list        List journal entries, with an optional starting datetime.
  search      Search journal entries using fuzzy search.
  timesince   Output a string of time since last post.

Flags:
  -h, --help      help for etu
  -v, --version   version for etu

Use "etu [command] --help" for more information about a command.
```

## Inspiration

Etu is [the personification of time](https://en.wikipedia.org/wiki/Time_and_fate_deities) according to the [Lakota](https://en.wikipedia.org/wiki/Lakota_people).

Etu is inspired heavily by the work of @neauoire at [wiki.xxiivv.com](https://wiki.xxiivv.com/#about), [Time Travelers](https://github.com/merveilles/Time-Travelers), and the screenshots in the [inspiration](https://github.com/icco/etu/tree/main/inspiration) folder.

Other projects that inspired me:

 - https://github.com/charmbracelet/glow
 - https://github.com/caarlos0/tasktimer
 - https://github.com/achannarasappa/ticker

## History

I've rewritten this approximately seven times. Originally a location-based blogging app, then time tracking, wiki, journaling. Backends: Charm, SQLite, Notion, and now the [etu-backend](https://github.com/icco/etu-backend) gRPC API.
