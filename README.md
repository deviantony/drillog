# drillog

Hierarchical logging for Go. Flat files, tree views.

## The Problem

Modern apps generate thousands of log lines. Concurrent operations interleave. Finding one request's journey through grep is painful.

## The Solution

Keep logs flat (greppable), but add minimal metadata. A viewer reconstructs the hierarchy on demand.

```
Your App  →  Flat Log File  →  Tree Viewer
              (still works      (see the
               with grep)        hierarchy)
```

## Installation

```bash
# Library
go get github.com/deviantony/drillog

# Viewer CLI
go install github.com/deviantony/drillog/cmd/drillog@latest
```

## Quick Start

```go
package main

import (
    "context"
    "os"
    "github.com/deviantony/drillog"
)

func main() {
    drillog.SetDefault(drillog.NewTextHandler(os.Stderr, nil))

    ctx := context.Background()
    ctx, end := drillog.Start(ctx, "main")
    defer end()

    drillog.Info(ctx, "starting")
    processOrder(ctx, 12345)
}

func processOrder(ctx context.Context, orderID int) {
    ctx, end := drillog.Start(ctx, "processOrder")
    defer end()

    drillog.Info(ctx, "loading order", "order_id", orderID)
}
```

**Output:**
```
time=2025-12-04T10:00:00Z level=INFO msg="main started" span=a1b2c3d4
time=2025-12-04T10:00:00Z level=INFO msg="starting" span=a1b2c3d4
time=2025-12-04T10:00:00Z level=INFO msg="processOrder started" span=e5f6a7b8 parent=a1b2c3d4
time=2025-12-04T10:00:00Z level=INFO msg="loading order" order_id=12345 span=e5f6a7b8 parent=a1b2c3d4
time=2025-12-04T10:00:00Z level=INFO msg="processOrder completed" duration=1.2ms span=e5f6a7b8 parent=a1b2c3d4
time=2025-12-04T10:00:00Z level=INFO msg="main completed" duration=2.5ms span=a1b2c3d4
```

## Viewer

View logs as an interactive tree in your browser:

```bash
# Run your app, capture logs
go run ./myapp 2> app.log

# Open viewer
drillog view app.log
```

The viewer auto-opens in your browser with:
- Collapsible span tree
- Search across messages and attributes
- Level filters (DEBUG/INFO/WARN/ERROR)
- Duration badges

**Flags:**
- `-port 8080` - Use specific port
- `-host 0.0.0.0` - Bind to all interfaces (for containers)
- `--no-browser` - Don't auto-open browser

## API Reference

### Spans

```go
ctx, end := drillog.Start(ctx, "operation")
defer end()
```

### Logging

```go
drillog.Debug(ctx, "message", "key", "value")
drillog.Info(ctx, "message", "key", "value")
drillog.Warn(ctx, "message", "key", "value")
drillog.Error(ctx, "message", "key", "value")
```

### Configuration

```go
// Text output
drillog.SetDefault(drillog.NewTextHandler(os.Stderr, nil))

// JSON output
drillog.SetDefault(drillog.NewJSONHandler(os.Stderr, nil))

// Wrap existing handler
drillog.SetDefault(drillog.NewHandler(myHandler, nil))
```

### Context Utilities

```go
spanID := drillog.SpanID(ctx)
parentID := drillog.ParentID(ctx)
```

## Design Principles

- **Zero external dependencies** - standard library only
- **Context-first** - no logger objects to pass around
- **Logs stay flat** - works with grep, tail, existing tools
- **Minimal API** - `Start()` + `defer end()` + log functions

## License

MIT
