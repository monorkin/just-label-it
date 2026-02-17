# CLAUDE.md

## Project

Just Label It (`jli`) — a Go CLI tool for labeling media files (images, video, audio) for LLM training sets. Single self-contained binary with an embedded web UI.

## Build

```bash
make build       # → ./bin/jli
make build-all   # → ./dist/jli-{os}-{arch}
make clean       # remove bin/ and dist/
make vet         # run go vet
```

Requires Go 1.25+. No CGO needed.

## Run

```bash
jli [directory]              # start server + open browser
jli serve [directory]        # start server only
jli --port 8080 .            # specify port
```

The database file `jli.db` is created in the target directory.

## Architecture

- **Module**: `github.com/monorkin/just-label-it`
- **CLI**: `cmd/` — Cobra commands (root defaults to `open`, plus `serve`)
- **Database**: `internal/db/` — SQLite via `modernc.org/sqlite`, WAL mode, single connection, migrations via `PRAGMA user_version`
- **Scanner**: `internal/scanner/` — walks directories, classifies files by extension
- **Server**: `internal/server/` — stdlib `net/http` routing (Go 1.25 method+wildcard), CSRF via `http.CrossOriginProtection`
- **Browser**: `internal/browser/` — opens URL with platform-native command
- **Web assets**: `web/` — embedded via `//go:embed`, HTML templates + Stimulus.js controllers

## Code conventions

- **Clarity over cleverness**. Small, focused functions. No premature abstraction.
- **Thin handlers, fat models**. HTTP handlers in `server/` parse input, call `db/`, render output. Business logic lives in `db/`.
- **Strict separation**. `db/` knows nothing about HTTP. `server/` knows nothing about SQL. `scanner/` is pure filesystem logic.
- **Error handling**. Always wrap with context: `fmt.Errorf("fetching media file %d: %w", id, err)`. Never swallow errors.
- **No interfaces** unless there are two implementations.
- **Frontend**: Stimulus.js (vendored UMD), no build step. Controllers are plain JS IIFEs that register on `window.StimulusApp`.

## Key design decisions

- Auto-save everything — no save button. Descriptions debounced 500ms client-side.
- Navigation ordered alphabetically by path with wrap-around.
- Video/audio files get a pinned 0:00 keyframe (cannot be moved or deleted). Protected in `db/keyframes.go` via `ErrPinnedKeyframe`.
- Media files served at `/media/{path...}` with path-traversal protection (absolute path prefix check).
- Labels are global and shared across files/keyframes. Created on first use.
