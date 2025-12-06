package viewer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestServer() *Server {
	now := time.Now()
	entries := []Entry{
		{Time: now, Level: "INFO", Message: "main started", Span: "aaa"},
		{Time: now.Add(time.Millisecond), Level: "DEBUG", Message: "debug info", Span: "aaa"},
		{Time: now.Add(2 * time.Millisecond), Level: "INFO", Message: "child started", Span: "bbb", Parent: "aaa"},
		{Time: now.Add(3 * time.Millisecond), Level: "WARN", Message: "slow query", Span: "bbb", Parent: "aaa", Attrs: map[string]string{"duration": "500ms"}},
		{Time: now.Add(4 * time.Millisecond), Level: "INFO", Message: "child completed", Span: "bbb", Parent: "aaa", Attrs: map[string]string{"duration": "2ms"}},
		{Time: now.Add(5 * time.Millisecond), Level: "ERROR", Message: "something failed", Span: "aaa"},
		{Time: now.Add(6 * time.Millisecond), Level: "INFO", Message: "main completed", Span: "aaa", Attrs: map[string]string{"duration": "6ms"}},
	}

	tree := BuildTree(entries)
	return NewServer(tree, entries)
}

func TestHandleTree(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/tree", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp TreeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check roots
	if len(resp.Roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(resp.Roots))
	}
	if resp.Roots[0] != "aaa" {
		t.Errorf("expected root 'aaa', got %s", resp.Roots[0])
	}

	// Check spans
	if len(resp.Spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(resp.Spans))
	}

	// Check root span
	root := resp.Spans["aaa"]
	if root.Name != "main" {
		t.Errorf("expected root name 'main', got %s", root.Name)
	}
	if len(root.Children) != 1 || root.Children[0] != "bbb" {
		t.Errorf("expected children [bbb], got %v", root.Children)
	}
	if root.LogCount != 4 {
		t.Errorf("expected 4 logs in root, got %d", root.LogCount)
	}

	// Check child span
	child := resp.Spans["bbb"]
	if child.Parent != "aaa" {
		t.Errorf("expected parent 'aaa', got %s", child.Parent)
	}
}

func TestHandleTree_MethodNotAllowed(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/tree", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandleLogs(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/logs?span=bbb", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp LogsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(resp.Logs))
	}

	// Check that logs have correct span
	for _, log := range resp.Logs {
		if log.Span != "bbb" {
			t.Errorf("expected span 'bbb', got %s", log.Span)
		}
	}
}

func TestHandleLogs_MissingSpan(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleLogs_SpanNotFound(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/logs?span=nonexistent", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleStats(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp StatsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.TotalSpans != 2 {
		t.Errorf("expected 2 spans, got %d", resp.TotalSpans)
	}
	if resp.TotalLogs != 7 {
		t.Errorf("expected 7 logs, got %d", resp.TotalLogs)
	}
	if resp.Levels["INFO"] != 4 {
		t.Errorf("expected 4 INFO, got %d", resp.Levels["INFO"])
	}
	if resp.Levels["DEBUG"] != 1 {
		t.Errorf("expected 1 DEBUG, got %d", resp.Levels["DEBUG"])
	}
	if resp.Levels["WARN"] != 1 {
		t.Errorf("expected 1 WARN, got %d", resp.Levels["WARN"])
	}
	if resp.Levels["ERROR"] != 1 {
		t.Errorf("expected 1 ERROR, got %d", resp.Levels["ERROR"])
	}
}

func TestHandleSearch(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=slow", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 match, got %d", resp.Total)
	}
	if len(resp.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(resp.Matches))
	}
	if resp.Matches[0].Message != "slow query" {
		t.Errorf("expected 'slow query', got %s", resp.Matches[0].Message)
	}
}

func TestHandleSearch_CaseInsensitive(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=FAILED", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 match for 'FAILED', got %d", resp.Total)
	}
}

func TestHandleSearch_SearchesAttrs(t *testing.T) {
	server := setupTestServer()

	// Search for "500ms" which is in attrs, not message
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=500ms", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 match for '500ms' in attrs, got %d", resp.Total)
	}
}

func TestHandleSearch_MissingQuery(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleSearch_NoMatches(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=nonexistent", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("expected 0 matches, got %d", resp.Total)
	}
}

func TestContentTypeJSON(t *testing.T) {
	server := setupTestServer()

	endpoints := []string{"/api/tree", "/api/logs?span=aaa", "/api/stats", "/api/search?q=main"}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("%s: expected Content-Type 'application/json', got %s", endpoint, contentType)
		}
	}
}

func TestEmptyTree(t *testing.T) {
	tree := BuildTree([]Entry{})
	server := NewServer(tree, []Entry{})

	// Test /api/tree with empty data
	req := httptest.NewRequest(http.MethodGet, "/api/tree", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp TreeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should return empty arrays, not null
	if resp.Roots == nil {
		t.Error("expected roots to be empty array, got nil")
	}
	if resp.Spans == nil {
		t.Error("expected spans to be empty object, got nil")
	}

	// Test /api/stats with empty data
	req = httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	var stats StatsResponse
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}

	if stats.TotalSpans != 0 || stats.TotalLogs != 0 {
		t.Errorf("expected 0 spans and logs, got %d and %d", stats.TotalSpans, stats.TotalLogs)
	}
}
