# Etu

[![Go Report Card](https://goreportcard.com/badge/github.com/icco/etu)](https://goreportcard.com/report/github.com/icco/etu)
[![Go Reference](https://pkg.go.dev/badge/go.icco.me/etu.svg)](https://pkg.go.dev/go.icco.me/etu)

Etu is a command line client for [Etu](https://etu.timeclimbers.com), a journal for [interstitial journaling](https://betterhumans.pub/replace-your-to-do-list-with-interstitial-journaling-to-increase-productivity-4e43109d15ef): short notes ("blips") written between tasks.

## Install

Homebrew (macOS and Linux):

```shell
brew tap icco/tap
brew install --cask etu
```

With Go 1.26 or later:

```shell
go install go.icco.me/etu@latest
```

Binaries for Linux, macOS, and Windows are on the [releases page](https://github.com/icco/etu/releases).

## Quickstart

1. Sign up at [etu.timeclimbers.com](https://etu.timeclimbers.com) and create an API key under [Settings > API](https://etu.timeclimbers.com/settings/api).
2. Run `etu login` and paste the key.
3. Run `etu create` to write your first blip, then `etu list` to read it back.

## Pricing

See [pricing](https://etu.timeclimbers.com/pricing) for current plans and trial terms. Journaling commands require subscription access; `export` remains available after a subscription expires or is canceled.

## Commands

| Command | Description |
| --- | --- |
| `create` | Write a new entry. Attach images with `-i` and audio with `-a`. |
| `list` | List entries, optionally from a starting datetime. |
| `show` | Show an entry with its images and audio. |
| `last` | Print the most recent entry. |
| `random` | Show a random entry. |
| `search` | Fuzzy search entries. |
| `edit` | Edit an entry. |
| `delete` | Delete an entry. |
| `tags` | List tags with usage counts. |
| `stats` | Show your blip, tag, and word counts. |
| `export` | Export all of your data as JSON to stdout or a file (`-o`). |
| `timesince` | Print the time since your last entry, for shell prompts. |
| `login` / `logout` | Save or remove your API key. |

Run `etu [command] --help` for flags.

## Configuration

`etu login` writes `~/.config/etu/config.json`:

```json
{
  "api_key": "etu_...",
  "grpc_target": "grpc.etu.timeclimbers.com:443"
}
```

`ETU_API_KEY` and `ETU_GRPC_TARGET` override the file.

## Build from source

Requires Go 1.26 or later.

```shell
git clone https://github.com/icco/etu
cd etu
go build -o etu .
```

## Inspiration

Etu is [the personification of time](https://en.wikipedia.org/wiki/Time_and_fate_deities) according to the [Lakota](https://en.wikipedia.org/wiki/Lakota_people). It is inspired by [wiki.xxiivv.com](https://wiki.xxiivv.com/#about), [Time Travelers](https://github.com/merveilles/Time-Travelers), [glow](https://github.com/charmbracelet/glow), [tasktimer](https://github.com/caarlos0/tasktimer), [ticker](https://github.com/achannarasappa/ticker), and the screenshots in [inspiration](https://github.com/icco/etu/tree/main/inspiration).
