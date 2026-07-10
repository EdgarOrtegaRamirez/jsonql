package output

import (
	"strings"
	"testing"
)

func TestJSONWriter_Pretty(t *testing.T) {
	data := map[string]interface{}{"name": "Alice", "age": 30}
	var buf strings.Builder
	w := NewJSONWriter(false)
	if err := w.Write(data, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "Alice"`) {
		t.Errorf("expected pretty JSON, got: %s", buf.String())
	}
}

func TestJSONWriter_Compact(t *testing.T) {
	data := map[string]interface{}{"name": "Alice", "age": 30}
	var buf strings.Builder
	w := NewJSONWriter(true)
	if err := w.Write(data, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Compact JSON should not have indentation (no "  " patterns)
	if strings.Contains(buf.String(), "  ") {
		t.Errorf("expected compact JSON without indentation, got: %s", buf.String())
	}
}

func TestJSONLWriter(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice"},
		map[string]interface{}{"name": "Bob"},
	}
	var buf strings.Builder
	w := NewJSONLWriter()
	if err := w.Write(data, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "Alice") {
		t.Errorf("expected Alice in first line, got: %s", lines[0])
	}
}

func TestTextWriter_NonArray(t *testing.T) {
	data := map[string]interface{}{"name": "Alice", "age": 30}
	var buf strings.Builder
	w := NewTextWriter()
	if err := w.Write(data, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), "Alice") {
		t.Errorf("expected Alice in output, got: %s", buf.String())
	}
}

func TestInfoWriter_Object(t *testing.T) {
	data := map[string]interface{}{
		"name": "Alice",
		"age":  30,
		"city": "NYC",
	}
	var buf strings.Builder
	w := NewInfoWriter()
	if err := w.Write(data, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Type: object") {
		t.Errorf("expected Type: object, got: %s", output)
	}
	if !strings.Contains(output, "Count: 3") {
		t.Errorf("expected Count: 3, got: %s", output)
	}
	if !strings.Contains(output, "Keys:") {
		t.Errorf("expected Keys: in output, got: %s", output)
	}
}

func TestInfoWriter_Array(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice"},
		map[string]interface{}{"name": "Bob"},
	}
	var buf strings.Builder
	w := NewInfoWriter()
	if err := w.Write(data, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Type: array") {
		t.Errorf("expected Type: array, got: %s", output)
	}
	if !strings.Contains(output, "Count: 2") {
		t.Errorf("expected Count: 2, got: %s", output)
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"json", 0},
		{"JSON", 0},
		{"jsonl", 1},
		{"JSONL", 1},
		{"text", 2},
		{"TEXT", 2},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := ParseFormat(tt.input)
		if got != Format(tt.want) {
			t.Errorf("ParseFormat(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
