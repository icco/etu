# CLI Feature Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the etu CLI to feature parity with the web and mobile clients by adding `edit`, `tags`, and `stats` commands.

**Architecture:** Follow the existing Cobra + client-package split: each command gets a thin cobra.Command in the main package that delegates to a method on `client.Config` (client package), which calls the backend over gRPC. TagsService and StatsService clients must be added to `client/grpc.go` alongside the existing NotesService/ApiKeysService clients.

**Tech Stack:** Go, Cobra, Charmbracelet (huh forms, bubbletea), gRPC + `github.com/icco/etu-backend` protos.

**Backend API shapes (from etu-backend/proto/etu.proto):**
- `TagsService.ListTags(ListTagsRequest{user_id}) → ListTagsResponse{repeated Tag tags}` where `Tag{id, name, count, created_at}`
- `StatsService.GetStats(GetStatsRequest{user_id}) → GetStatsResponse{total_blips, unique_tags, words_written}` — empty `user_id` returns global stats; set `user_id` for per-user stats.
- `NotesService.UpdateNote(UpdateNoteRequest{user_id, id, optional content, repeated tags, update_tags bool, add_images, add_audios}) → UpdateNoteResponse{note}`

---

### Task 1: Tags client method

**Files:**
- Modify: `client/grpc.go` (add `tags proto.TagsServiceClient` next to existing service clients, initialized in the same place as `notes`)
- Modify: `client/client.go` (add `ListTags` method)
- Test: `client/grpc_test.go` (conversion helpers if any)

- [ ] Add `TagsServiceClient` to the lazily-initialized gRPC client set in `client/grpc.go`, mirroring how `proto.NewNotesServiceClient(conn)` is created.
- [ ] Add to `client/client.go`:

```go
// Tag is a user tag with its usage count.
type Tag struct {
	Name  string
	Count int32
}

// ListTags returns all tags for the authenticated user, sorted by the backend.
func (c *Config) ListTags(ctx context.Context) ([]Tag, error) {
	userID, err := c.userID(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.grpcTags().ListTags(ctx, &proto.ListTagsRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	tags := make([]Tag, 0, len(resp.GetTags()))
	for _, t := range resp.GetTags() {
		tags = append(tags, Tag{Name: t.GetName(), Count: t.GetCount()})
	}
	return tags, nil
}
```

(Adapt the user-ID lookup and client accessor names to the existing patterns in grpc.go — read how ListPosts resolves userID first.)

- [ ] Add a unit test for any pure conversion logic; run `go test ./...` — PASS.
- [ ] Commit: `feat: add ListTags client method`

### Task 2: `etu tags` command

**Files:**
- Modify: `main.go` (register command)
- Create: `tags.go`
- Test: `main_test.go` or `tags_test.go` for any formatting helper

- [ ] Create `tags.go` with a `tagsCmd` that prints each tag as `name (count)` one per line (plain output suits piping; no TUI needed). Sort by count descending, then name.
- [ ] Register in `main.go` `rootCmd.AddCommand(...)` next to the other commands, alias `t`.
- [ ] Test formatting helper; run `go test ./...` — PASS. Commit: `feat: add tags command`

### Task 3: Stats client method + `etu stats` command

**Files:**
- Modify: `client/grpc.go` (StatsService client), `client/client.go` (GetStats), `main.go`
- Create: `stats.go`

- [ ] Add `StatsServiceClient` wiring as in Task 1.
- [ ] Add to `client/client.go`:

```go
// Stats holds aggregate journal statistics.
type Stats struct {
	TotalBlips   int64
	UniqueTags   int64
	WordsWritten int64
}

// GetStats returns stats for the authenticated user; pass global=true for community stats.
func (c *Config) GetStats(ctx context.Context, global bool) (Stats, error) { ... }
```

with the same userID/empty-userID split described in the API shapes above.
- [ ] Create `stats.go` defining `statsCmd`: prints `Blips: N`, `Tags: N`, `Words written: N`; `--global` flag adds community stats. Register in main.go.
- [ ] `go test ./...` PASS, `go vet ./...` clean. Commit: `feat: add stats command`

### Task 4: `etu edit` command

**Files:**
- Modify: `client/client.go` (UpdatePost), `client/grpc.go` (nothing new — NotesService exists), `main.go`
- Create: `edit.go`

- [ ] Add `UpdatePost(ctx, pageID string, content string) (*Post, error)` to client package calling `UpdateNote` with `Content: &content` and `UpdateTags: false` (tags unchanged; backend re-derives nothing). Wrap errors like the neighbors.
- [ ] Create `edit.go`: reuse the entry-selection list used by `show`/`delete` (see show.go:20-43 and main.go:208-257) to pick an entry, fetch full content via the existing `GetPostFullContent`, pre-fill a `huh` multiline text field with the current content, and save via `UpdatePost`. Alias `e`. Abort cleanly if content unchanged.
- [ ] `go test ./...` PASS. Commit: `feat: add edit command`

### Task 5: Docs

- [ ] Update README.md usage section with the three new commands; regenerate completions is handled by the build (Taskfile `completions` target — note it, don't run if binary build unavailable).
- [ ] Create `CLAUDE.md` (~30 lines): what the repo is, build/test/lint commands (`task build`, `go test ./...`, `go vet ./...`), architecture (main package = cobra/TUI; client package = gRPC; protos from etu-backend), config location (`~/.config/etu/config.json`, env overrides `ETU_API_KEY`/`ETU_GRPC_TARGET`).
- [ ] Commit: `docs: document new commands, add CLAUDE.md`

### Verification

- [ ] `go build ./...`, `go test ./...`, `go vet ./...` all pass.
- [ ] `staticcheck ./...` if available.
