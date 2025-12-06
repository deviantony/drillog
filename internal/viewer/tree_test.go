package viewer

import (
	"strings"
	"testing"
	"time"
)

func TestBuildTree_SimpleHierarchy(t *testing.T) {
	// Root span with one child
	entries := []Entry{
		{Time: time.Now(), Level: "INFO", Message: "main started", Span: "aaa"},
		{Time: time.Now(), Level: "INFO", Message: "doing work", Span: "aaa"},
		{Time: time.Now(), Level: "INFO", Message: "child started", Span: "bbb", Parent: "aaa"},
		{Time: time.Now(), Level: "INFO", Message: "child completed", Span: "bbb", Parent: "aaa", Attrs: map[string]string{"duration": "10ms"}},
		{Time: time.Now(), Level: "INFO", Message: "main completed", Span: "aaa", Attrs: map[string]string{"duration": "50ms"}},
	}

	tree := BuildTree(entries)

	// Should have 1 root
	if len(tree.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree.Roots))
	}
	if tree.Roots[0] != "aaa" {
		t.Errorf("expected root 'aaa', got %s", tree.Roots[0])
	}

	// Should have 2 spans
	if len(tree.Spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(tree.Spans))
	}

	// Check root span
	root := tree.Spans["aaa"]
	if root.Name != "main" {
		t.Errorf("expected root name 'main', got %s", root.Name)
	}
	if root.Duration != "50ms" {
		t.Errorf("expected root duration '50ms', got %s", root.Duration)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0] != "bbb" {
		t.Errorf("expected child 'bbb', got %s", root.Children[0])
	}

	// Check child span
	child := tree.Spans["bbb"]
	if child.Name != "child" {
		t.Errorf("expected child name 'child', got %s", child.Name)
	}
	if child.Parent != "aaa" {
		t.Errorf("expected child parent 'aaa', got %s", child.Parent)
	}
	if child.Duration != "10ms" {
		t.Errorf("expected child duration '10ms', got %s", child.Duration)
	}
}

func TestBuildTree_MultipleRoots(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{Time: now, Level: "INFO", Message: "first started", Span: "aaa"},
		{Time: now.Add(time.Second), Level: "INFO", Message: "second started", Span: "bbb"},
		{Time: now.Add(2 * time.Second), Level: "INFO", Message: "first completed", Span: "aaa"},
		{Time: now.Add(3 * time.Second), Level: "INFO", Message: "second completed", Span: "bbb"},
	}

	tree := BuildTree(entries)

	if len(tree.Roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(tree.Roots))
	}

	// Roots should be sorted by start time
	if tree.Roots[0] != "aaa" || tree.Roots[1] != "bbb" {
		t.Errorf("expected roots sorted as [aaa, bbb], got %v", tree.Roots)
	}
}

func TestBuildTree_DeepNesting(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{Time: now, Level: "INFO", Message: "level1 started", Span: "l1"},
		{Time: now.Add(time.Millisecond), Level: "INFO", Message: "level2 started", Span: "l2", Parent: "l1"},
		{Time: now.Add(2 * time.Millisecond), Level: "INFO", Message: "level3 started", Span: "l3", Parent: "l2"},
		{Time: now.Add(3 * time.Millisecond), Level: "INFO", Message: "level3 completed", Span: "l3", Parent: "l2"},
		{Time: now.Add(4 * time.Millisecond), Level: "INFO", Message: "level2 completed", Span: "l2", Parent: "l1"},
		{Time: now.Add(5 * time.Millisecond), Level: "INFO", Message: "level1 completed", Span: "l1"},
	}

	tree := BuildTree(entries)

	// Verify hierarchy: l1 -> l2 -> l3
	l1 := tree.Spans["l1"]
	if len(l1.Children) != 1 || l1.Children[0] != "l2" {
		t.Errorf("l1 should have child l2, got %v", l1.Children)
	}

	l2 := tree.Spans["l2"]
	if len(l2.Children) != 1 || l2.Children[0] != "l3" {
		t.Errorf("l2 should have child l3, got %v", l2.Children)
	}

	l3 := tree.Spans["l3"]
	if len(l3.Children) != 0 {
		t.Errorf("l3 should have no children, got %v", l3.Children)
	}
}

func TestBuildTree_OrphanedSpan(t *testing.T) {
	entries := []Entry{
		{Time: time.Now(), Level: "INFO", Message: "orphan started", Span: "orphan", Parent: "nonexistent"},
		{Time: time.Now(), Level: "INFO", Message: "orphan completed", Span: "orphan", Parent: "nonexistent"},
	}

	tree := BuildTree(entries)

	// Orphaned span should become a root
	if len(tree.Roots) != 1 {
		t.Fatalf("expected 1 root (orphan), got %d", len(tree.Roots))
	}
	if tree.Roots[0] != "orphan" {
		t.Errorf("expected orphan as root, got %s", tree.Roots[0])
	}
}

func TestBuildTree_ChildrenSortedByTime(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{Time: now, Level: "INFO", Message: "parent started", Span: "parent"},
		// Children added in reverse order
		{Time: now.Add(3 * time.Second), Level: "INFO", Message: "child3 started", Span: "c3", Parent: "parent"},
		{Time: now.Add(1 * time.Second), Level: "INFO", Message: "child1 started", Span: "c1", Parent: "parent"},
		{Time: now.Add(2 * time.Second), Level: "INFO", Message: "child2 started", Span: "c2", Parent: "parent"},
	}

	tree := BuildTree(entries)

	parent := tree.Spans["parent"]
	if len(parent.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(parent.Children))
	}

	// Children should be sorted by start time
	expected := []string{"c1", "c2", "c3"}
	for i, want := range expected {
		if parent.Children[i] != want {
			t.Errorf("child %d: expected %s, got %s", i, want, parent.Children[i])
		}
	}
}

func TestBuildTree_Stats(t *testing.T) {
	entries := []Entry{
		{Time: time.Now(), Level: "INFO", Message: "main started", Span: "a"},
		{Time: time.Now(), Level: "DEBUG", Message: "debug msg", Span: "a"},
		{Time: time.Now(), Level: "INFO", Message: "child started", Span: "b", Parent: "a"},
		{Time: time.Now(), Level: "WARN", Message: "warning", Span: "b", Parent: "a"},
		{Time: time.Now(), Level: "ERROR", Message: "error", Span: "b", Parent: "a"},
	}

	tree := BuildTree(entries)
	stats := tree.Stats()

	if stats.TotalSpans != 2 {
		t.Errorf("expected 2 spans, got %d", stats.TotalSpans)
	}
	if stats.TotalLogs != 5 {
		t.Errorf("expected 5 logs, got %d", stats.TotalLogs)
	}
	if stats.Levels["INFO"] != 2 {
		t.Errorf("expected 2 INFO, got %d", stats.Levels["INFO"])
	}
	if stats.Levels["DEBUG"] != 1 {
		t.Errorf("expected 1 DEBUG, got %d", stats.Levels["DEBUG"])
	}
	if stats.Levels["WARN"] != 1 {
		t.Errorf("expected 1 WARN, got %d", stats.Levels["WARN"])
	}
	if stats.Levels["ERROR"] != 1 {
		t.Errorf("expected 1 ERROR, got %d", stats.Levels["ERROR"])
	}
}

func TestBuildTree_EmptyEntries(t *testing.T) {
	tree := BuildTree([]Entry{})

	if len(tree.Roots) != 0 {
		t.Errorf("expected 0 roots, got %d", len(tree.Roots))
	}
	if len(tree.Spans) != 0 {
		t.Errorf("expected 0 spans, got %d", len(tree.Spans))
	}
}

func TestBuildTree_EntriesWithoutSpan(t *testing.T) {
	entries := []Entry{
		{Time: time.Now(), Level: "INFO", Message: "no span here"},
		{Time: time.Now(), Level: "INFO", Message: "has span", Span: "aaa"},
	}

	tree := BuildTree(entries)

	// Only the entry with span should be included
	if len(tree.Spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tree.Spans))
	}
	if _, exists := tree.Spans["aaa"]; !exists {
		t.Error("expected span 'aaa' to exist")
	}
}

func TestBuildTree_Integration(t *testing.T) {
	// Parse real log output and build tree
	input := `time=2025-12-04T10:00:00Z level=INFO msg="sync-cycle started" span=root123
time=2025-12-04T10:00:01Z level=INFO msg="starting device sync" span=root123
time=2025-12-04T10:00:02Z level=INFO msg="fetch-agents started" span=child1 parent=root123
time=2025-12-04T10:00:03Z level=INFO msg="agents retrieved" count=5 span=child1 parent=root123
time=2025-12-04T10:00:04Z level=INFO msg="fetch-agents completed" duration=2s span=child1 parent=root123
time=2025-12-04T10:00:05Z level=INFO msg="process-device started" span=child2 parent=root123
time=2025-12-04T10:00:06Z level=INFO msg="sync-data started" span=grandchild parent=child2
time=2025-12-04T10:00:07Z level=INFO msg="sync-data completed" duration=1s span=grandchild parent=child2
time=2025-12-04T10:00:08Z level=INFO msg="process-device completed" duration=3s span=child2 parent=root123
time=2025-12-04T10:00:09Z level=INFO msg="sync-cycle completed" duration=9s span=root123`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tree := BuildTree(result.Entries)

	// Verify structure
	if len(tree.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree.Roots))
	}

	root := tree.Spans["root123"]
	if root.Name != "sync-cycle" {
		t.Errorf("expected root name 'sync-cycle', got %s", root.Name)
	}
	if root.Duration != "9s" {
		t.Errorf("expected root duration '9s', got %s", root.Duration)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}

	// Check grandchild is nested correctly
	child2 := tree.Spans["child2"]
	if len(child2.Children) != 1 || child2.Children[0] != "grandchild" {
		t.Errorf("expected child2 to have grandchild, got %v", child2.Children)
	}

	// Check stats
	stats := tree.Stats()
	if stats.TotalSpans != 4 {
		t.Errorf("expected 4 spans, got %d", stats.TotalSpans)
	}
	if stats.TotalLogs != 10 {
		t.Errorf("expected 10 logs, got %d", stats.TotalLogs)
	}
}

func TestIsStartedMessage(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"main started", true},
		{"process-device started", true},
		{"started", true},
		{"main completed", false},
		{"starting up", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isStartedMessage(tt.msg)
		if got != tt.want {
			t.Errorf("isStartedMessage(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

func TestIsCompletedMessage(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"main completed", true},
		{"process-device completed", true},
		{"completed", true},
		{"main started", false},
		{"completing task", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isCompletedMessage(tt.msg)
		if got != tt.want {
			t.Errorf("isCompletedMessage(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

func TestExtractSpanName(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"main started", "main"},
		{"process-device started", "process-device"},
		{"sync-cycle started", "sync-cycle"},
		{"started", "started"},
		{"no suffix", "no suffix"},
	}

	for _, tt := range tests {
		got := extractSpanName(tt.msg)
		if got != tt.want {
			t.Errorf("extractSpanName(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}
