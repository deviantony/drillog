package viewer

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Server serves the viewer REST API and static UI.
type Server struct {
	tree    *Tree
	entries []Entry
	mux     *http.ServeMux
}

// NewServer creates a new viewer server with the given tree and entries.
func NewServer(tree *Tree, entries []Entry) *Server {
	s := &Server{
		tree:    tree,
		entries: entries,
		mux:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/tree", s.handleTree)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/search", s.handleSearch)

	// Serve embedded UI for all other routes
	uiHandler := UIHandler()
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// For SPA routing: serve index.html for non-asset paths
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/assets/") {
			r.URL.Path = "/"
		}
		uiHandler.ServeHTTP(w, r)
	})
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// TreeResponse is the JSON response for GET /api/tree.
type TreeResponse struct {
	Roots []string                `json:"roots"`
	Spans map[string]SpanResponse `json:"spans"`
}

// SpanResponse is the JSON representation of a span.
type SpanResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Parent    string   `json:"parent,omitempty"`
	Children  []string `json:"children"`
	StartTime string   `json:"startTime,omitempty"`
	Duration  string   `json:"duration,omitempty"`
	LogCount  int      `json:"logCount"`
}

// handleTree handles GET /api/tree.
func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := TreeResponse{
		Roots: s.tree.Roots,
		Spans: make(map[string]SpanResponse),
	}

	for id, span := range s.tree.Spans {
		sr := SpanResponse{
			ID:       span.ID,
			Name:     span.Name,
			Parent:   span.Parent,
			Children: span.Children,
			Duration: span.Duration,
			LogCount: len(span.Entries),
		}
		if !span.StartTime.IsZero() {
			sr.StartTime = span.StartTime.Format("2006-01-02T15:04:05.999Z07:00")
		}
		// Ensure children is never null in JSON
		if sr.Children == nil {
			sr.Children = []string{}
		}
		resp.Spans[id] = sr
	}

	// Ensure roots is never null in JSON
	if resp.Roots == nil {
		resp.Roots = []string{}
	}

	s.writeJSON(w, resp)
}

// LogsResponse is the JSON response for GET /api/logs.
type LogsResponse struct {
	Logs []LogEntry `json:"logs"`
}

// LogEntry is the JSON representation of a log entry.
type LogEntry struct {
	Time    string            `json:"time"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Span    string            `json:"span"`
	Parent  string            `json:"parent,omitempty"`
	Attrs   map[string]string `json:"attrs,omitempty"`
}

// handleLogs handles GET /api/logs?span={spanId}.
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	spanID := r.URL.Query().Get("span")
	if spanID == "" {
		http.Error(w, "span parameter required", http.StatusBadRequest)
		return
	}

	span, exists := s.tree.Spans[spanID]
	if !exists {
		http.Error(w, "span not found", http.StatusNotFound)
		return
	}

	resp := LogsResponse{
		Logs: make([]LogEntry, 0, len(span.Entries)),
	}

	for _, e := range span.Entries {
		le := LogEntry{
			Level:   e.Level,
			Message: e.Message,
			Span:    e.Span,
			Parent:  e.Parent,
		}
		if !e.Time.IsZero() {
			le.Time = e.Time.Format("2006-01-02T15:04:05.999Z07:00")
		}
		if len(e.Attrs) > 0 {
			le.Attrs = e.Attrs
		}
		resp.Logs = append(resp.Logs, le)
	}

	s.writeJSON(w, resp)
}

// StatsResponse is the JSON response for GET /api/stats.
type StatsResponse struct {
	TotalSpans int            `json:"totalSpans"`
	TotalLogs  int            `json:"totalLogs"`
	Levels     map[string]int `json:"levels"`
}

// handleStats handles GET /api/stats.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.tree.Stats()
	resp := StatsResponse{
		TotalSpans: stats.TotalSpans,
		TotalLogs:  stats.TotalLogs,
		Levels:     stats.Levels,
	}

	// Ensure levels is never null
	if resp.Levels == nil {
		resp.Levels = make(map[string]int)
	}

	s.writeJSON(w, resp)
}

// SearchResponse is the JSON response for GET /api/search.
type SearchResponse struct {
	Matches []LogEntry `json:"matches"`
	Total   int        `json:"total"`
}

// handleSearch handles GET /api/search?q={query}.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}

	query = strings.ToLower(query)
	matches := make([]LogEntry, 0)

	for _, e := range s.entries {
		if matchesQuery(e, query) {
			le := LogEntry{
				Level:   e.Level,
				Message: e.Message,
				Span:    e.Span,
				Parent:  e.Parent,
			}
			if !e.Time.IsZero() {
				le.Time = e.Time.Format("2006-01-02T15:04:05.999Z07:00")
			}
			if len(e.Attrs) > 0 {
				le.Attrs = e.Attrs
			}
			matches = append(matches, le)
		}
	}

	resp := SearchResponse{
		Matches: matches,
		Total:   len(matches),
	}

	s.writeJSON(w, resp)
}

// matchesQuery checks if an entry matches the search query.
// Searches message and attribute values (case-insensitive).
func matchesQuery(e Entry, query string) bool {
	if strings.Contains(strings.ToLower(e.Message), query) {
		return true
	}
	for _, v := range e.Attrs {
		if strings.Contains(strings.ToLower(v), query) {
			return true
		}
	}
	return false
}

// writeJSON writes a JSON response with appropriate headers.
func (s *Server) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
