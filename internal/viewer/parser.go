// Package viewer provides log parsing and tree building for the drillog viewer.
package viewer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Entry represents a parsed log entry.
type Entry struct {
	Time    time.Time
	Level   string
	Message string
	Span    string
	Parent  string
	Attrs   map[string]string
}

// Format represents the detected log format.
type Format int

const (
	FormatUnknown Format = iota
	FormatText
	FormatJSON
)

// ParseResult contains all parsed entries and metadata.
type ParseResult struct {
	Entries []Entry
	Format  Format
}

// Parse reads log lines from r and returns parsed entries.
func Parse(r io.Reader) (*ParseResult, error) {
	scanner := bufio.NewScanner(r)
	result := &ParseResult{
		Entries: make([]Entry, 0),
		Format:  FormatUnknown,
	}

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Detect format from first non-empty line
		if result.Format == FormatUnknown {
			if strings.HasPrefix(line, "{") {
				result.Format = FormatJSON
			} else {
				result.Format = FormatText
			}
		}

		var entry Entry
		var err error

		if result.Format == FormatJSON {
			entry, err = parseJSONLine(line)
		} else {
			entry, err = parseTextLine(line)
		}

		if err != nil {
			// Skip malformed lines but continue parsing
			continue
		}

		// Validate entry has minimum required fields
		if !isValidEntry(entry) {
			continue
		}

		result.Entries = append(result.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return result, nil
}

// jsonEntry is the structure for JSON log lines from slog.JSONHandler.
type jsonEntry struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"msg"`
	Span    string         `json:"span"`
	Parent  string         `json:"parent"`
	Extra   map[string]any `json:"-"`
}

func parseJSONLine(line string) (Entry, error) {
	var entry Entry

	// First unmarshal into a map to get all fields
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return entry, err
	}

	// Extract known fields
	if t, ok := raw["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			entry.Time = parsed
		}
	}
	if l, ok := raw["level"].(string); ok {
		entry.Level = l
	}
	if m, ok := raw["msg"].(string); ok {
		entry.Message = m
	}
	if s, ok := raw["span"].(string); ok {
		entry.Span = s
	}
	if p, ok := raw["parent"].(string); ok {
		entry.Parent = p
	}

	// Collect remaining attributes
	entry.Attrs = make(map[string]string)
	knownKeys := map[string]bool{"time": true, "level": true, "msg": true, "span": true, "parent": true}
	for k, v := range raw {
		if !knownKeys[k] {
			entry.Attrs[k] = fmt.Sprintf("%v", v)
		}
	}

	return entry, nil
}

func parseTextLine(line string) (Entry, error) {
	var entry Entry
	entry.Attrs = make(map[string]string)

	// slog.TextHandler format: time=... level=... msg="..." key=value ...
	// Fields are space-separated, values may be quoted

	pairs, err := parseKeyValuePairs(line)
	if err != nil {
		return entry, err
	}

	for key, value := range pairs {
		switch key {
		case "time":
			if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
				entry.Time = parsed
			}
		case "level":
			entry.Level = value
		case "msg":
			entry.Message = value
		case "span":
			entry.Span = value
		case "parent":
			entry.Parent = value
		default:
			entry.Attrs[key] = value
		}
	}

	return entry, nil
}

// parseKeyValuePairs parses a line of key=value pairs.
// Values may be quoted with double quotes if they contain spaces.
func parseKeyValuePairs(line string) (map[string]string, error) {
	result := make(map[string]string)
	i := 0
	n := len(line)

	for i < n {
		// Skip whitespace
		for i < n && line[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}

		// Parse key (until '=')
		keyStart := i
		for i < n && line[i] != '=' && line[i] != ' ' {
			i++
		}
		if i >= n || line[i] != '=' {
			// No '=' found, skip this token
			for i < n && line[i] != ' ' {
				i++
			}
			continue
		}
		key := line[keyStart:i]
		i++ // skip '='

		if i >= n {
			result[key] = ""
			break
		}

		// Parse value
		var value string
		if line[i] == '"' {
			// Quoted value
			i++ // skip opening quote
			valueStart := i
			for i < n && line[i] != '"' {
				if line[i] == '\\' && i+1 < n {
					i += 2 // skip escaped character
				} else {
					i++
				}
			}
			value = unescapeQuoted(line[valueStart:i])
			if i < n {
				i++ // skip closing quote
			}
		} else {
			// Unquoted value (until space)
			valueStart := i
			for i < n && line[i] != ' ' {
				i++
			}
			value = line[valueStart:i]
		}

		result[key] = value
	}

	return result, nil
}

// isValidEntry checks if an entry has the minimum required fields.
// A valid entry must have at least a level and message.
func isValidEntry(e Entry) bool {
	return e.Level != "" && e.Message != ""
}

// unescapeQuoted handles escape sequences in quoted strings.
func unescapeQuoted(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte(s[i+1])
			}
			i += 2
		} else {
			b.WriteByte(s[i])
			i++
		}
	}

	return b.String()
}
