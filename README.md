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

## Usage

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
2025-12-04T10:00:00Z INFO main started span=a1b2c3d4
2025-12-04T10:00:00Z INFO starting span=a1b2c3d4
2025-12-04T10:00:00Z INFO processOrder started span=e5f6a7b8 parent=a1b2c3d4
2025-12-04T10:00:00Z INFO loading order order_id=12345 span=e5f6a7b8 parent=a1b2c3d4
2025-12-04T10:00:00Z INFO processOrder completed duration=1.2ms span=e5f6a7b8 parent=a1b2c3d4
2025-12-04T10:00:00Z INFO main completed duration=2.5ms span=a1b2c3d4
```

**In the viewer:**
```
▼ main [2.5ms]
  │ starting
  ▼ processOrder [1.2ms]
    │ loading order order_id=12345
```

## Installation

```bash
go get github.com/deviantony/drillog
```

## API

### Starting Spans

```go
ctx, end := drillog.Start(ctx, "operation name")
defer end()
```

Nested calls automatically capture parent relationships.

### Logging

```go
drillog.Debug(ctx, "message", "key", "value")
drillog.Info(ctx, "message", "key", "value")
drillog.Warn(ctx, "message", "key", "value")
drillog.Error(ctx, "message", "key", "value")
```

Or use standard slog (if handler is configured):

```go
slog.InfoContext(ctx, "message", "key", "value")
```

### Configuration

```go
// Text output (human-readable)
drillog.SetDefault(drillog.NewTextHandler(os.Stderr, nil))

// JSON output
drillog.SetDefault(drillog.NewJSONHandler(os.Stderr, nil))

// Wrap existing handler
drillog.SetDefault(drillog.NewHandler(myHandler, nil))

// Custom ID generator
drillog.SetDefault(drillog.NewTextHandler(os.Stderr, &drillog.HandlerOptions{
    IDGenerator: myIDFunc,
}))
```

### Context Utilities

```go
spanID := drillog.SpanID(ctx)     // current span ID
parentID := drillog.ParentID(ctx) // parent span ID
```

## Design Principles

- **Zero config works** - defaults to `slog.Default()`
- **Context-first** - no logger objects to pass around
- **Standard library only** - no external dependencies
- **Logs stay flat** - works with grep, tail, existing tools
- **Minimal API** - `Start()` + `defer end()` + log functions

## Viewer

Open `viewer.html` in any browser. Drag and drop your log file. See the tree.

## License

MIT
