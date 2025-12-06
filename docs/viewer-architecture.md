# Viewer Architecture

The drillog viewer is a CLI tool that parses log files and displays them as an interactive tree in the browser.

## Overview

```
┌─────────────────────────────────────────────────────┐
│                    Go Binary                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │ Log Parser  │  │ HTTP Server │  │ Embedded UI │ │
│  │ (startup)   │→ │ (REST API)  │→ │ (static)    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────┘
         ↓                  ↓
    Builds tree       Browser auto-opens
    in memory         http://localhost:PORT
```

## CLI Usage

```bash
drillog view app.log                    # Auto-opens browser
drillog view -port 8080 app.log         # Specific port
drillog view -host 0.0.0.0 app.log      # Bind all interfaces
drillog view --no-browser app.log       # Don't auto-open
```

## Project Structure

```
drillog/
├── cmd/drillog/main.go        # CLI entrypoint
├── internal/viewer/
│   ├── parser.go              # Log file parser
│   ├── tree.go                # Tree building logic
│   ├── server.go              # HTTP server + REST API
│   └── embed.go               # go:embed for UI assets
└── ui/                        # Frontend source (Svelte)
    ├── src/
    │   ├── App.svelte
    │   └── components/
    │       ├── Tree.svelte
    │       ├── TreeNode.svelte
    │       ├── LogEntry.svelte
    │       ├── Search.svelte
    │       ├── Filters.svelte
    │       └── Stats.svelte
    └── dist/                  # Built frontend (embedded)
```

## REST API

### GET /api/tree

Returns tree structure with spans.

```json
{
  "roots": ["span1"],
  "spans": {
    "span1": {
      "id": "span1",
      "name": "main",
      "parent": "",
      "children": ["span2"],
      "startTime": "2025-12-04T10:00:00Z",
      "duration": "2.5s",
      "logCount": 5
    }
  }
}
```

### GET /api/logs?span={spanId}

Returns log entries for a span.

```json
{
  "logs": [
    {
      "time": "2025-12-04T10:00:00Z",
      "level": "INFO",
      "message": "Starting",
      "span": "span1",
      "attrs": {"key": "value"}
    }
  ]
}
```

### GET /api/stats

Returns aggregate statistics.

```json
{
  "totalSpans": 100,
  "totalLogs": 500,
  "levels": {"DEBUG": 50, "INFO": 400, "WARN": 40, "ERROR": 10}
}
```

### GET /api/search?q={query}

Searches messages and attributes (case-insensitive).

```json
{
  "matches": [...],
  "total": 42
}
```

## Log Parsing

### Supported Formats

**Text (slog.TextHandler):**
```
time=2025-12-04T10:00:00Z level=INFO msg="message" span=abc123 parent=def456
```

**JSON (slog.JSONHandler):**
```json
{"time":"2025-12-04T10:00:00Z","level":"INFO","msg":"message","span":"abc123","parent":"def456"}
```

Format is auto-detected from the first line.

### Tree Building

1. Parse all log lines into entries
2. Group entries by span ID
3. Extract span name from "started" messages, duration from "completed" messages
4. Link children to parents via `parent` field
5. Orphaned spans (parent not found) become roots
6. Sort by start time

## Tech Stack

| Component | Technology |
|-----------|------------|
| CLI | `flag` (stdlib) |
| HTTP | `net/http` (stdlib) |
| Frontend | Svelte + TailwindCSS |
| Build | Vite |
| Embedding | `go:embed` |

## Development

### Frontend

```bash
cd ui
npm install
npm run dev      # Dev server
npm run build    # Production build
```

### Building CLI

```bash
cd ui && npm run build && cd ..
cp -r ui/dist internal/viewer/ui/
go build ./cmd/drillog
```
