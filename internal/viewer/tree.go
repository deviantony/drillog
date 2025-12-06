package viewer

import (
	"sort"
	"time"
)

// Span represents a span in the tree with its log entries.
type Span struct {
	ID        string
	Name      string
	Parent    string
	Children  []string
	StartTime time.Time
	Duration  string
	Entries   []Entry
}

// Tree represents the reconstructed log hierarchy.
type Tree struct {
	Roots []string          // Span IDs of root nodes (no parent)
	Spans map[string]*Span  // All spans indexed by ID
}

// BuildTree constructs a tree from parsed log entries.
func BuildTree(entries []Entry) *Tree {
	tree := &Tree{
		Roots: make([]string, 0),
		Spans: make(map[string]*Span),
	}

	if len(entries) == 0 {
		return tree
	}

	// Pass 1: Group entries by span ID
	for _, e := range entries {
		if e.Span == "" {
			continue
		}

		span, exists := tree.Spans[e.Span]
		if !exists {
			span = &Span{
				ID:       e.Span,
				Parent:   e.Parent,
				Children: make([]string, 0),
				Entries:  make([]Entry, 0),
			}
			tree.Spans[e.Span] = span
		}

		// Update parent if not set (first entry with parent wins)
		if span.Parent == "" && e.Parent != "" {
			span.Parent = e.Parent
		}

		// Extract span metadata from "started" and "completed" messages
		if isStartedMessage(e.Message) {
			span.Name = extractSpanName(e.Message)
			if span.StartTime.IsZero() {
				span.StartTime = e.Time
			}
		}
		if isCompletedMessage(e.Message) {
			if d, ok := e.Attrs["duration"]; ok {
				span.Duration = d
			}
		}

		span.Entries = append(span.Entries, e)
	}

	// Pass 2: Build parent-child relationships and identify roots
	for spanID, span := range tree.Spans {
		if span.Parent == "" {
			// No parent = root node
			tree.Roots = append(tree.Roots, spanID)
		} else if parentSpan, exists := tree.Spans[span.Parent]; exists {
			// Link to parent
			parentSpan.Children = append(parentSpan.Children, spanID)
		} else {
			// Orphaned span (parent not found) = treat as root
			tree.Roots = append(tree.Roots, spanID)
		}
	}

	// Sort roots and children by start time
	tree.sortByStartTime()

	return tree
}

// sortByStartTime sorts roots and all children by their start time.
func (t *Tree) sortByStartTime() {
	// Sort roots
	sort.Slice(t.Roots, func(i, j int) bool {
		si, sj := t.Spans[t.Roots[i]], t.Spans[t.Roots[j]]
		return si.StartTime.Before(sj.StartTime)
	})

	// Sort children of each span
	for _, span := range t.Spans {
		if len(span.Children) > 1 {
			sort.Slice(span.Children, func(i, j int) bool {
				ci, cj := t.Spans[span.Children[i]], t.Spans[span.Children[j]]
				return ci.StartTime.Before(cj.StartTime)
			})
		}
	}
}

// Stats returns aggregate statistics about the tree.
func (t *Tree) Stats() TreeStats {
	stats := TreeStats{
		Levels: make(map[string]int),
	}

	for _, span := range t.Spans {
		stats.TotalSpans++
		stats.TotalLogs += len(span.Entries)

		for _, e := range span.Entries {
			stats.Levels[e.Level]++
		}
	}

	return stats
}

// TreeStats contains aggregate statistics.
type TreeStats struct {
	TotalSpans int
	TotalLogs  int
	Levels     map[string]int
}

// isStartedMessage checks if a message indicates span start.
func isStartedMessage(msg string) bool {
	if msg == "started" {
		return true
	}
	n := len(msg)
	if n < 8 {
		return false
	}
	return msg[n-8:] == " started"
}

// isCompletedMessage checks if a message indicates span completion.
func isCompletedMessage(msg string) bool {
	if msg == "completed" {
		return true
	}
	n := len(msg)
	if n < 10 {
		return false
	}
	return msg[n-10:] == " completed"
}

// extractSpanName extracts the span name from a "started" message.
// "my-span started" → "my-span"
func extractSpanName(msg string) string {
	n := len(msg)
	if n <= 8 {
		return msg
	}
	if msg[n-8:] == " started" {
		return msg[:n-8]
	}
	return msg
}
