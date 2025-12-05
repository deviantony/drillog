# Viewer Architecture

This document describes the architecture of the drillog viewer, a CLI tool that parses log files and displays them as an interactive tree in the browser.

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

## Design Decisions

### Use Case

- **Post-mortem analysis**: Analyze completed log files, not real-time tailing
- **Local development**: Runs on developer's machine, no remote/CI requirements
- **Large files**: Must handle 10MB+ logs with hundreds of thousands of lines

### Why Go CLI + Web UI?

| Consideration | Decision |
|---------------|----------|
| Performance | Go parses logs faster than JavaScript |
| Large files | Go can handle arbitrarily large files in memory |
| Distribution | Single binary via `go install` |
| UI richness | Web UI enables interactive tree, search, filters |
| Simplicity | Auto-opens browser, no manual setup |

### Data Loading Strategy

**Load on startup**: The entire log file is parsed when the CLI starts. This provides faster UI interactions after the initial load at the cost of higher memory usage. This trade-off is acceptable for local development use.

## CLI Usage

```bash
# Install
go install github.com/deviantony/drillog/cmd/drillog@latest

# View a log file (auto-opens browser)
drillog view app.log
```

Output:
```
Parsing app.log...
Loaded 150,000 log entries, 12,000 spans
Serving on http://localhost:8374
```

## Tech Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| CLI framework | `flag` (stdlib) | Simple, no dependencies |
| HTTP server | `net/http` (stdlib) | No dependencies |
| Frontend | Svelte + TailwindCSS | Small bundle (~15KB), fast, clean DX |
| Virtual scrolling | Custom or `svelte-virtual-list` | Required for 100k+ nodes |
| Build tool | Vite | Fast builds, good Svelte support |
| Embedding | `go:embed` | Single binary distribution |

## Project Structure

```
drillog/
├── drillog.go                 # Library (existing)
├── cmd/
│   └── drillog/
│       └── main.go            # CLI entrypoint
├── internal/
│   └── viewer/
│       ├── server.go          # HTTP server + handlers
│       ├── parser.go          # Log file parser
│       ├── tree.go            # Tree building logic
│       └── embed.go           # go:embed for UI assets
├── ui/                        # Frontend source (Svelte)
│   ├── src/
│   │   ├── App.svelte
│   │   ├── components/
│   │   │   ├── Tree.svelte
│   │   │   ├── TreeNode.svelte
│   │   │   ├── LogEntry.svelte
│   │   │   ├── Search.svelte
│   │   │   └── Filters.svelte
│   │   └── lib/
│   │       └── api.js
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── internal/viewer/ui/dist/   # Built frontend (embedded)
└── docs/
    └── viewer-architecture.md # This document
```

## REST API

### GET /api/tree

Returns the complete tree structure with spans and their relationships.

```json
{
  "roots": ["span1", "span2"],
  "spans": {
    "span1": {
      "id": "span1",
      "name": "main",
      "parent": null,
      "children": ["span2", "span3"],
      "startTime": "2025-12-04T10:00:00Z",
      "duration": "2.5s",
      "logCount": 5
    }
  }
}
```

### GET /api/logs?span={spanId}

Returns all log entries for a specific span.

```json
{
  "logs": [
    {
      "time": "2025-12-04T10:00:00Z",
      "level": "INFO",
      "message": "Starting application",
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
  "totalSpans": 12000,
  "totalLogs": 150000,
  "totalDuration": "45.2s",
  "levels": {
    "DEBUG": 50000,
    "INFO": 80000,
    "WARN": 15000,
    "ERROR": 5000
  }
}
```

### GET /api/search?q={query}

Returns log entries matching the search query.

```json
{
  "matches": [
    {
      "time": "2025-12-04T10:00:05Z",
      "level": "ERROR",
      "message": "Connection failed",
      "span": "span5",
      "attrs": {"error": "timeout"}
    }
  ],
  "total": 42
}
```

## UI Features

### Tree View
- Collapsible nodes with expand/collapse all
- Virtual scrolling for performance (render only visible nodes)
- Indentation with tree connectors (├─, └─, │)
- Expand/collapse indicators (▶, ▼)

### Visual Design
- Dark theme with terminal aesthetic
- Monospace font throughout
- Color-coded log levels:
  - DEBUG: blue
  - INFO: green
  - WARN: orange
  - ERROR: red
- Duration badges on span nodes
- Dimmed metadata (timestamps, span IDs)

### Interactions
- Click to expand/collapse spans
- Search box with real-time filtering
- Level filter toggles (show/hide DEBUG, INFO, WARN, ERROR)
- Keyboard navigation (future enhancement)

### Performance Targets
- Initial render: < 500ms for 100k logs
- Expand/collapse: < 50ms
- Search: < 200ms
- Smooth scrolling at 60fps

## Log Parsing

### Supported Formats

**Text format (slog TextHandler):**
```
time=2025-12-04T10:00:00Z level=INFO msg="message" span=abc123 parent=def456
```

**JSON format (slog JSONHandler):**
```json
{"time":"2025-12-04T10:00:00Z","level":"INFO","msg":"message","span":"abc123","parent":"def456"}
```

### Parser Requirements

1. Auto-detect format (text vs JSON) from first line
2. Extract: timestamp, level, message, span, parent, other attributes
3. Handle malformed lines gracefully (skip or mark as unparseable)
4. Build tree structure from span/parent relationships
5. Handle orphaned spans (parent not found) by attaching to root

### Tree Building Algorithm

```
1. Parse all log lines into entries
2. Group entries by span ID
3. For each unique span:
   a. Find the "started" entry (extract name, start time)
   b. Find the "completed" entry (extract duration)
   c. Record parent ID
4. Build tree:
   a. Spans with no parent → root nodes
   b. Spans with parent → children of that parent
5. Sort children by start time
```

## Build & Development

### Frontend Development

```bash
cd ui
npm install
npm run dev      # Dev server with hot reload
npm run build    # Build for production → dist/
```

### Embedding Frontend

```go
//go:embed ui/dist/*
var uiFS embed.FS
```

### Building the CLI

```bash
# Build frontend first
cd ui && npm run build && cd ..

# Build Go binary
go build -o drillog ./cmd/drillog
```

### Development Workflow

1. Run `npm run dev` in `ui/` for frontend development
2. Run Go server separately pointing to dev server (proxy mode)
3. For production: build frontend, then build Go binary

