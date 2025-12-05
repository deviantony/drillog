# LogDrill: Project Requirements

## What You're Building

A Go logging library that makes debugging easy:
- Logs stay **flat** (works with grep/tail)
- A viewer shows them as **trees** (hierarchical debugging)
- Dead simple API: `logger.Start(ctx, "name")` + `defer end()`

**The Problem:** Modern apps generate thousands of log lines. Concurrent operations interleave. Finding one request's journey through the system is painful with just grep.

**The Solution:** Keep logs flat (greppable), but add minimal metadata that allows a smart viewer to reconstruct the execution hierarchy on demand.

**Target Users:** Solo developers and small teams who want powerful debugging without infrastructure complexity.

---

## System Architecture

```
┌─────────────────┐
│   Your Go App   │
│  logger.Start() │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Flat Log File  │  ← Still greppable!
│  span=abc123    │
│  parent=def456  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  HTML Viewer    │  ← Drag & drop file
│  Tree Structure │     See hierarchy
└─────────────────┘
```

**Key Insight:** Logs contain correlation IDs (`span`, `parent`). The viewer reconstructs the tree from these IDs.

**Philosophy:** Don't change how logs work. Just add smart tooling to view them.

---

## Component 1: Go Logging Library

### What It Must Do

**Core Functionality:**
- Wrap Go's standard `log/slog` package
- Add automatic span correlation via `context.Context`
- Generate unique span IDs automatically
- Track parent-child relationships between spans
- Output logs in a format that's both human-readable and machine-parseable

**Technical Requirements:**
- Zero external dependencies (use Go stdlib only)
- Works with Go 1.21+
- Thread-safe for concurrent operations
- Minimal performance overhead
- Compatible with standard Go patterns

### Developer Experience (How Developers Will Use It)

**Basic Usage:**
```go
package main

import (
    "context"
    "log/slog"
    "github.com/yourorg/logdrill/logger"
)

func main() {
    ctx := context.Background()
    
    // Start a span - automatic span ID generation
    ctx, end := logger.Start(ctx, "Main Workflow")
    defer end()
    
    // Regular logging - span IDs added automatically
    slog.InfoContext(ctx, "Starting application")
    
    // Nested spans create hierarchy
    processOrder(ctx)
}

func processOrder(ctx context.Context) {
    // Child span - parent ID captured from context
    ctx, end := logger.Start(ctx, "Process Order")
    defer end()
    
    slog.InfoContext(ctx, "Loading order data")
    slog.InfoContext(ctx, "Order processed", "order_id", 12345)
}
```

**Expected Log Output:**
```
2025-12-04T10:00:00Z INFO Main Workflow started span=a1b2c3d4
2025-12-04T10:00:01Z INFO Starting application span=a1b2c3d4
2025-12-04T10:00:02Z INFO Process Order started span=b2c3d4e5 parent=a1b2c3d4
2025-12-04T10:00:03Z INFO Loading order data span=b2c3d4e5 parent=a1b2c3d4
2025-12-04T10:00:04Z INFO Order processed order_id=12345 span=b2c3d4e5 parent=a1b2c3d4
2025-12-04T10:00:05Z INFO Process Order completed duration=3.2s span=b2c3d4e5 parent=a1b2c3d4
2025-12-04T10:00:05Z INFO Main Workflow completed duration=5.1s span=a1b2c3d4
```

### Key Design Decisions

**Span IDs:**
- Must be short (8 characters) for readability
- Must be unique enough to avoid collisions in reasonable usage
- Hex format for easy parsing

**Context Propagation:**
- Use `context.Context` to carry span information
- Idiomatic Go pattern
- Works with existing Go code

**Automatic Start/End:**
- `Start()` logs the beginning and returns an end function
- Developer uses `defer end()` for automatic completion logging
- End function captures duration automatically

**Integration with slog:**
- Should work seamlessly with standard `slog` calls
- Developers can use either `logger.Info(ctx, ...)` convenience functions OR standard `slog.InfoContext(ctx, ...)`
- Both should include span metadata

---

## Component 2: HTML Viewer

### What It Must Do

**Core Functionality:**
- Accept log files via drag-and-drop or file picker
- Parse log lines and extract metadata (timestamp, level, message, span, parent)
- Reconstruct hierarchical tree structure from flat logs
- Display logs as collapsible/expandable tree
- Provide search and filtering capabilities
- Show timing and statistics

**Technical Requirements:**
- Single HTML file (fully self-contained)
- No external dependencies or internet connection required
- Works in any modern browser
- Handles log files up to 10MB reasonably well
- Client-side processing only (no backend needed)

### User Experience (How Users Will Use It)

**Basic Workflow:**
1. User runs their Go application: `go run main.go 2> app.log`
2. User opens `viewer.html` in any browser
3. User drags `app.log` onto the page (or clicks to browse)
4. Viewer instantly displays logs as a tree

**Visual Tree Structure:**
```
▼ Main Workflow started [5.1s]
  │ Starting application
  ▼ Process Order started [3.2s]
    │ Loading order data
    │ Order processed order_id=12345
```

**Interactive Features:**
- Click arrows to expand/collapse nodes
- Search box to find specific text
- Filter buttons to show/hide log levels (INFO, DEBUG, WARN, ERROR)
- Statistics panel showing: total spans, total logs, duration, error count
- Keyboard shortcuts for navigation

### Parsing Requirements

**The viewer must parse these log formats:**

**Console format:**
```
2025-12-04T10:00:00Z INFO message span=abc123 parent=def456
```

**JSON format (future):**
```json
{"time":"2025-12-04T10:00:00Z","level":"INFO","msg":"message","span":"abc123","parent":"def456"}
```

**Metadata to Extract:**
- `timestamp` - ISO 8601 format
- `level` - INFO, DEBUG, WARN, ERROR
- `message` - The log message text
- `span` - 8-character hex ID
- `parent` - 8-character hex ID (optional)
- `duration` - For "completed" messages (e.g., "2.5s", "150ms")

### Tree Reconstruction Algorithm

**High-Level Approach:**

1. **Parse Phase:** Read all log lines and extract metadata
2. **Grouping Phase:** Group logs by their `span` ID
3. **Linking Phase:** Create parent-child relationships using `parent` field
4. **Rendering Phase:** Display as nested HTML elements

**Key Rules:**
- Logs with the same `span` ID belong together
- If log has `parent=xyz`, it's a child of the span with `span=xyz`
- Logs without a `parent` field are root nodes
- Orphaned logs (parent not found) should still display

**Example:**
```
Log: span=aaa                    → Root node
Log: span=bbb parent=aaa         → Child of aaa
Log: span=ccc parent=aaa         → Child of aaa
Log: span=ddd parent=bbb         → Child of bbb (grandchild of aaa)
```

Results in tree:
```
aaa
├── bbb
│   └── ddd
└── ccc
```

### Visual Design Requirements

**Color Coding:**
- INFO logs: Green
- DEBUG logs: Blue
- WARN logs: Orange/Yellow
- ERROR logs: Red

**Typography:**
- Monospace font (like terminal)
- Dim colors for metadata (timestamps, span IDs)
- Bold or emphasized for messages
- Clear visual hierarchy

**Layout:**
- Tree connectors (├─, └─, │) for hierarchy
- Indentation for nesting levels
- Expand/collapse indicators (▶, ▼)
- Compact by default, comfortable spacing

---

## Component 3: Example Application

### What It Must Include

A working example that demonstrates:
- Basic span creation
- Nested spans (2-3 levels deep)
- Multiple concurrent operations
- Different log levels
- Various log messages with structured data

**Example Scenario:**
A device sync operation that:
1. Starts main sync cycle
2. Fetches agent list
3. Processes multiple devices in parallel
4. Shows some successes and some warnings

This proves the library works and shows users how to use it.

---

## Project Structure

```
logdrill/
├── logger/
│   └── logger.go           # Core library
├── examples/
│   └── device-sync/
│       └── main.go         # Working example
├── viewer.html             # Single-file viewer
├── go.mod
├── go.sum
└── README.md               # Usage documentation
```

---

## Success Criteria

The project is complete when:

### Library Success Criteria
- ✅ Developer can call `logger.Start(ctx, name)` and get automatic span tracking
- ✅ Nested spans automatically capture parent relationships
- ✅ Logs contain `span=` and `parent=` fields
- ✅ Works with standard `slog` logging
- ✅ Zero external dependencies
- ✅ Context is the only parameter needed

### Viewer Success Criteria
- ✅ User can drag-and-drop log file onto page
- ✅ Tree structure is correctly reconstructed
- ✅ Nested operations appear as nested nodes
- ✅ Search highlights matching logs
- ✅ Level filters work correctly
- ✅ Statistics panel shows accurate data
- ✅ Expand/collapse works smoothly

### Documentation Success Criteria
- ✅ README explains what LogDrill is and why it exists
- ✅ README shows basic usage example
- ✅ README explains how to run example
- ✅ README shows expected output
- ✅ Code comments explain key functions

---

## Design Principles

### For the Library

**1. Minimal API Surface**
- One main function: `Start()`
- Optional convenience functions
- No configuration objects or complex setup

**2. Context-First**
- Everything flows through `context.Context`
- No global state
- No logger objects to pass around

**3. Zero Ceremony**
- No manual ID management
- No explicit parent linking
- Just `Start()` and `defer end()`

**4. Standard Library**
- Build on `log/slog`
- Use Go idioms
- Zero external dependencies

### For the Viewer

**1. Zero Installation**
- Single HTML file
- No build step
- No npm/yarn/webpack
- Works offline

**2. Instant Feedback**
- Drag & drop = immediate tree view
- No loading spinners for reasonable files
- Smooth interactions

**3. Progressive Disclosure**
- Start collapsed, expand what you care about
- Hide DEBUG logs by default
- Search reveals relevant sections

**4. Familiar UX**
- Terminal aesthetic (monospace, colors)
- Tree connectors like `tree` command
- Keyboard shortcuts like vim/less

---

## Non-Requirements (Out of Scope)

These are explicitly NOT part of this project:

❌ Real-time log streaming (logs must be saved to file first)
❌ Multi-service distributed tracing (single service only)
❌ Log aggregation or storage (no database)
❌ Alerting or monitoring (just viewing)
❌ Authentication or access control
❌ Cloud hosting or SaaS offering
❌ Mobile app or native desktop viewer
❌ Integration with observability platforms
❌ Custom log formats beyond console/JSON

**Why?** Keep it simple. Solve one problem well: local debugging with hierarchical log viewing.

---

## Questions to Consider During Implementation

### Library Questions
- How do you generate unique span IDs efficiently?
- How do you store span info in context without collisions?
- How do you handle nil context gracefully?
- How do you ensure thread safety for concurrent spans?
- What happens if a span never calls `end()`?

### Viewer Questions
- How do you handle malformed log lines?
- What if a log references a parent that doesn't exist?
- How do you handle out-of-order logs?
- How do you efficiently search large log files?
- How do you prevent browser freezing on huge files?

### UX Questions
- Should everything be expanded or collapsed by default?
- How do you indicate that a span has children?
- How do you show timing information clearly?
- Should search auto-expand parent nodes?
- What keyboard shortcuts are most useful?

---

## Testing Strategy

### Library Testing
- Unit tests for span ID generation
- Tests for context propagation
- Tests for nested spans
- Tests for concurrent operations
- Integration test with real logging

### Viewer Testing
- Test with valid log files
- Test with malformed logs
- Test with missing parent spans
- Test with very large files (stress test)
- Test search functionality
- Test filters
- Cross-browser testing

### End-to-End Testing
- Run example application
- Generate logs
- Open in viewer
- Verify tree structure matches code structure
- Verify all features work

---

## Timeline Estimate

**Week 1:**
- Day 1-2: Library implementation
- Day 3-4: Viewer implementation
- Day 5: Example application

**Week 2:**
- Day 1-2: Testing and bug fixes
- Day 3: Documentation
- Day 4-5: Polish and refinement

**Total:** 2 weeks for complete, polished implementation

---

## Resources

- Go slog package: https://pkg.go.dev/log/slog
- Go context package: https://pkg.go.dev/context
- Existing prototype in provided files (for reference, not to copy)

---

## Getting Help

If you need clarification on:
- **Requirements:** Ask before implementing
- **Design decisions:** Propose your approach for feedback
- **Technical challenges:** Discuss alternatives

The goal is a simple, elegant solution. When in doubt, choose simplicity.
