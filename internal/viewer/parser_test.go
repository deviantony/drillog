package viewer

import (
	"strings"
	"testing"
	"time"
)

func TestParse_TextFormat(t *testing.T) {
	input := `time=2025-12-04T10:00:00Z level=INFO msg="main started" span=a1b2c3d4
time=2025-12-04T10:00:01Z level=INFO msg="processing" span=b2c3d4e5 parent=a1b2c3d4
time=2025-12-04T10:00:02Z level=INFO msg="main completed" duration=2s span=a1b2c3d4`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Format != FormatText {
		t.Errorf("expected FormatText, got %v", result.Format)
	}

	if len(result.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entries))
	}

	// Check first entry
	e := result.Entries[0]
	if e.Level != "INFO" {
		t.Errorf("entry 0: expected level INFO, got %s", e.Level)
	}
	if e.Message != "main started" {
		t.Errorf("entry 0: expected message 'main started', got %s", e.Message)
	}
	if e.Span != "a1b2c3d4" {
		t.Errorf("entry 0: expected span a1b2c3d4, got %s", e.Span)
	}
	if e.Parent != "" {
		t.Errorf("entry 0: expected no parent, got %s", e.Parent)
	}

	// Check second entry (has parent)
	e = result.Entries[1]
	if e.Span != "b2c3d4e5" {
		t.Errorf("entry 1: expected span b2c3d4e5, got %s", e.Span)
	}
	if e.Parent != "a1b2c3d4" {
		t.Errorf("entry 1: expected parent a1b2c3d4, got %s", e.Parent)
	}

	// Check third entry (has duration attribute)
	e = result.Entries[2]
	if e.Attrs["duration"] != "2s" {
		t.Errorf("entry 2: expected duration=2s, got %s", e.Attrs["duration"])
	}
}

func TestParse_JSONFormat(t *testing.T) {
	input := `{"time":"2025-12-04T10:00:00Z","level":"INFO","msg":"main started","span":"a1b2c3d4"}
{"time":"2025-12-04T10:00:01Z","level":"WARN","msg":"slow query","span":"b2c3d4e5","parent":"a1b2c3d4","query_ms":150}`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Format != FormatJSON {
		t.Errorf("expected FormatJSON, got %v", result.Format)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}

	// Check first entry
	e := result.Entries[0]
	if e.Level != "INFO" {
		t.Errorf("entry 0: expected level INFO, got %s", e.Level)
	}
	if e.Message != "main started" {
		t.Errorf("entry 0: expected message 'main started', got %s", e.Message)
	}
	if e.Span != "a1b2c3d4" {
		t.Errorf("entry 0: expected span a1b2c3d4, got %s", e.Span)
	}

	// Check second entry
	e = result.Entries[1]
	if e.Level != "WARN" {
		t.Errorf("entry 1: expected level WARN, got %s", e.Level)
	}
	if e.Parent != "a1b2c3d4" {
		t.Errorf("entry 1: expected parent a1b2c3d4, got %s", e.Parent)
	}
	if e.Attrs["query_ms"] != "150" {
		t.Errorf("entry 1: expected query_ms=150, got %s", e.Attrs["query_ms"])
	}
}

func TestParse_EmptyLines(t *testing.T) {
	input := `
time=2025-12-04T10:00:00Z level=INFO msg="test" span=abc123

time=2025-12-04T10:00:01Z level=DEBUG msg="another" span=def456
`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestParse_MalformedLines(t *testing.T) {
	input := `time=2025-12-04T10:00:00Z level=INFO msg="valid" span=abc123
this is not a valid log line
time=2025-12-04T10:00:01Z level=INFO msg="also valid" span=def456`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should skip malformed line and parse the valid ones
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestParse_QuotedValuesWithSpaces(t *testing.T) {
	input := `time=2025-12-04T10:00:00Z level=INFO msg="hello world with spaces" span=abc123 user="john doe"`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}

	e := result.Entries[0]
	if e.Message != "hello world with spaces" {
		t.Errorf("expected message 'hello world with spaces', got %s", e.Message)
	}
	if e.Attrs["user"] != "john doe" {
		t.Errorf("expected user='john doe', got %s", e.Attrs["user"])
	}
}

func TestParse_TimeParsing(t *testing.T) {
	input := `time=2025-12-04T10:30:45.123456789Z level=INFO msg="test" span=abc123`

	result, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	e := result.Entries[0]
	expected := time.Date(2025, 12, 4, 10, 30, 45, 123456789, time.UTC)
	if !e.Time.Equal(expected) {
		t.Errorf("expected time %v, got %v", expected, e.Time)
	}
}

func TestParse_AllLogLevels(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, level := range levels {
		input := `time=2025-12-04T10:00:00Z level=` + level + ` msg="test" span=abc123`
		result, err := Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("Parse failed for level %s: %v", level, err)
		}
		if result.Entries[0].Level != level {
			t.Errorf("expected level %s, got %s", level, result.Entries[0].Level)
		}
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "simple",
			input: "key=value",
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name:  "multiple",
			input: "a=1 b=2 c=3",
			expected: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
		},
		{
			name:  "quoted",
			input: `msg="hello world"`,
			expected: map[string]string{
				"msg": "hello world",
			},
		},
		{
			name:  "mixed",
			input: `level=INFO msg="hello world" count=42`,
			expected: map[string]string{
				"level": "INFO",
				"msg":   "hello world",
				"count": "42",
			},
		},
		{
			name:  "escaped quotes",
			input: `msg="say \"hello\""`,
			expected: map[string]string{
				"msg": `say "hello"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseKeyValuePairs(tt.input)
			if err != nil {
				t.Fatalf("parseKeyValuePairs failed: %v", err)
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("key %s: expected %q, got %q", k, v, result[k])
				}
			}
		})
	}
}
