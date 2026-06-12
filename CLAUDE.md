# CLAUDE.md

Etu is a personal command-line journaling tool (interstitial journaling). It talks gRPC to
[etu-backend](https://github.com/icco/etu-backend), whose generated protos are imported from
`github.com/icco/etu-backend/proto`.

## Commands

- `task build` — build the binary (or `go build -o etu .`)
- `go test ./...` — run tests
- `go vet ./...` — vet
- `task lint` — lint (runs `go vet ./...`)

## Architecture

Two packages:

- **main** (repo root): cobra commands (`main.go`, plus one file per larger command:
  `edit.go`, `stats.go`, `tags.go`, `show.go`, `search.go`) and the TUI. List selection is a
  bubbletea model (`list.go`); forms and confirmations use huh; long calls wrap in
  `huh/spinner`. Plain (pipeable) output is preferred when stdout is not a terminal.
- **client**: the gRPC client. `client/grpc.go` holds the lazily-initialized connection and
  service clients (Notes, ApiKeys, Tags, Stats) plus proto→display-model converters;
  `client/client.go` holds high-level methods (`ListPosts`, `SaveEntry`, `UpdatePost`,
  `ListTags`, `GetStats`, ...). The user ID is resolved once via `VerifyApiKey` and cached.

## Configuration

- Config file: `~/.config/etu/config.json` (`api_key`, `grpc_target`).
- Env overrides: `ETU_API_KEY`, `ETU_GRPC_TARGET`.
- The `timesince` command caches its result at `~/.config/etu/timesince.cache` (gob, 5 min TTL).
