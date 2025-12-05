// Package drillog provides hierarchical logging for Go.
// Logs stay flat (greppable), but contain span metadata that allows
// viewers to reconstruct the execution hierarchy.
package drillog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"sync"
	"time"
)

// IDGenerator generates span IDs.
type IDGenerator func() string

// HandlerOptions configures handler behavior.
type HandlerOptions struct {
	// IDGenerator generates span IDs. If nil, uses crypto/rand.
	IDGenerator IDGenerator
	// Level sets the minimum log level. If nil, defaults to INFO.
	Level slog.Leveler
}

// Handler is an slog.Handler that injects span attributes from context.
type Handler struct {
	inner slog.Handler
	idGen IDGenerator
}

// spanInfo holds span metadata stored in context.
type spanInfo struct {
	spanID   string
	parentID string
}

// contextKey is the type for context keys to avoid collisions.
type contextKey struct{}

var (
	spanKey = contextKey{}

	defaultHandler *Handler
	defaultMu      sync.RWMutex
)

// defaultIDGenerator generates 8-character hex IDs using crypto/rand.
func defaultIDGenerator() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to zero ID on error (shouldn't happen in practice)
		return "00000000"
	}
	return hex.EncodeToString(b)
}

// NewHandler wraps an existing slog.Handler with span injection.
func NewHandler(inner slog.Handler, opts *HandlerOptions) *Handler {
	h := &Handler{
		inner: inner,
		idGen: defaultIDGenerator,
	}
	if opts != nil && opts.IDGenerator != nil {
		h.idGen = opts.IDGenerator
	}
	return h
}

// NewTextHandler creates a Handler with a TextHandler writing to w.
func NewTextHandler(w io.Writer, opts *HandlerOptions) *Handler {
	var slogOpts *slog.HandlerOptions
	if opts != nil && opts.Level != nil {
		slogOpts = &slog.HandlerOptions{Level: opts.Level}
	}
	return NewHandler(slog.NewTextHandler(w, slogOpts), opts)
}

// NewJSONHandler creates a Handler with a JSONHandler writing to w.
func NewJSONHandler(w io.Writer, opts *HandlerOptions) *Handler {
	var slogOpts *slog.HandlerOptions
	if opts != nil && opts.Level != nil {
		slogOpts = &slog.HandlerOptions{Level: opts.Level}
	}
	return NewHandler(slog.NewJSONHandler(w, slogOpts), opts)
}

// SetDefault sets the default handler used by Start() and logging functions.
func SetDefault(h *Handler) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultHandler = h
}

// Default returns the current default handler.
// If none is set, returns nil (functions will use slog.Default()).
func Default() *Handler {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultHandler
}

// getHandler returns the handler to use for logging.
func getHandler() *Handler {
	if h := Default(); h != nil {
		return h
	}
	return nil
}

// getSlogLogger returns the slog.Logger to use for logging.
func getSlogLogger() *slog.Logger {
	if h := getHandler(); h != nil {
		return slog.New(h)
	}
	return slog.Default()
}

// Enabled implements slog.Handler.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	if info := getSpanInfo(ctx); info != nil {
		r.AddAttrs(slog.String("span", info.spanID))
		if info.parentID != "" {
			r.AddAttrs(slog.String("parent", info.parentID))
		}
	}
	return h.inner.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		inner: h.inner.WithAttrs(attrs),
		idGen: h.idGen,
	}
}

// WithGroup implements slog.Handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		inner: h.inner.WithGroup(name),
		idGen: h.idGen,
	}
}

// getSpanInfo retrieves span info from context.
func getSpanInfo(ctx context.Context) *spanInfo {
	if ctx == nil {
		return nil
	}
	if info, ok := ctx.Value(spanKey).(*spanInfo); ok {
		return info
	}
	return nil
}

// Start begins a new span and returns a context with span info and an end function.
// The end function logs completion with duration.
func Start(ctx context.Context, name string) (context.Context, func()) {
	if ctx == nil {
		ctx = context.Background()
	}

	h := getHandler()
	idGen := defaultIDGenerator
	if h != nil {
		idGen = h.idGen
	}

	// Generate new span ID
	spanID := idGen()

	// Get parent from existing context
	var parentID string
	if parent := getSpanInfo(ctx); parent != nil {
		parentID = parent.spanID
	}

	// Create new span info
	info := &spanInfo{
		spanID:   spanID,
		parentID: parentID,
	}

	// Store in context
	ctx = context.WithValue(ctx, spanKey, info)

	// Capture start time
	startTime := time.Now()

	// Log start
	// If drillog handler is set, it adds span/parent from context.
	// Otherwise, we add them manually for slog.Default() compatibility.
	logger := getSlogLogger()
	if getHandler() != nil {
		logger.InfoContext(ctx, name+" started")
	} else {
		if parentID != "" {
			logger.InfoContext(ctx, name+" started", "span", spanID, "parent", parentID)
		} else {
			logger.InfoContext(ctx, name+" started", "span", spanID)
		}
	}

	// Return end function
	end := func() {
		duration := time.Since(startTime)
		if getHandler() != nil {
			logger.InfoContext(ctx, name+" completed", "duration", formatDuration(duration))
		} else {
			if parentID != "" {
				logger.InfoContext(ctx, name+" completed", "duration", formatDuration(duration), "span", spanID, "parent", parentID)
			} else {
				logger.InfoContext(ctx, name+" completed", "duration", formatDuration(duration), "span", spanID)
			}
		}
	}

	return ctx, end
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(10 * time.Millisecond).String()
}

// SpanID returns the current span ID from context, or empty string if none.
func SpanID(ctx context.Context) string {
	if info := getSpanInfo(ctx); info != nil {
		return info.spanID
	}
	return ""
}

// ParentID returns the parent span ID from context, or empty string if none.
func ParentID(ctx context.Context) string {
	if info := getSpanInfo(ctx); info != nil {
		return info.parentID
	}
	return ""
}

// log is a helper that logs at the given level.
// If a drillog Handler is configured, span info is added by the handler.
// Otherwise, span info is added manually to support slog.Default().
func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger := getSlogLogger()
	if !logger.Enabled(ctx, level) {
		return
	}

	// Only add span info manually if no drillog handler (handler would do it)
	if getHandler() == nil {
		if info := getSpanInfo(ctx); info != nil {
			args = append(args, "span", info.spanID)
			if info.parentID != "" {
				args = append(args, "parent", info.parentID)
			}
		}
	}

	logger.Log(ctx, level, msg, args...)
}

// Debug logs at DEBUG level with span attributes from context.
func Debug(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, args...)
}

// Info logs at INFO level with span attributes from context.
func Info(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, args...)
}

// Warn logs at WARN level with span attributes from context.
func Warn(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, msg, args...)
}

// Error logs at ERROR level with span attributes from context.
func Error(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
}
