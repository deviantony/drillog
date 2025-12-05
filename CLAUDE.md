# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

- `go build ./...` - Build the library
- `go run ./examples/device-sync/` - Run the device sync example
- `go run ./examples/json-output/` - Run the JSON output example
- `go run ./examples/zero-config/` - Run the zero-config example

## Code Style

- Use Go standard formatting (`gofmt`)
- Zero external dependencies for the library - standard library only
- Keep the library as a single file (`drillog.go`)
- Use unexported types for context keys to avoid collisions

## Architecture

### Library (`drillog.go`)

Single-file library built on Go's `log/slog`:

- **Handler**: Wraps any `slog.Handler`, injects `span`/`parent` attributes from context
- **Context propagation**: Span info stored via `context.WithValue`
- **ID generation**: Pluggable `IDGenerator`, defaults to 8-char hex via `crypto/rand`
- **Dual-mode logging**: When drillog Handler is set, it auto-injects span attributes. When using `slog.Default()`, convenience functions manually append them.

### Viewer (`cmd/drillog`)

Go CLI that parses log files and serves an interactive web UI. See @docs/viewer-architecture.md for details.
