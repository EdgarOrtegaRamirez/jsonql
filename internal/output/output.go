// Package output provides formatters for JSON query results.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format represents an output format.
type Format int

const (
	JSON Format = iota
	JSONL
	TEXT
)

// ParseFormat returns the Format for a given string.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return JSON
	case "jsonl":
		return JSONL
	case "text":
		return TEXT
	default:
		return JSON
	}
}

// Writer is an interface for output formatters.
type Writer interface {
	Write(data interface{}, w io.Writer) error
}

// JSONWriter outputs data as formatted JSON.
type JSONWriter struct {
	Compact bool
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(compact bool) *JSONWriter {
	return &JSONWriter{Compact: compact}
}

func (w *JSONWriter) Write(data interface{}, out io.Writer) error {
	var buf []byte
	var err error
	if w.Compact {
		buf, err = json.Marshal(data)
	} else {
		buf, err = json.MarshalIndent(data, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	buf = append(buf, '\n')
	_, err = out.Write(buf)
	return err
}

// JSONLWriter outputs each element of an array as a JSON line.
type JSONLWriter struct{}

// NewJSONLWriter creates a new JSONL writer.
func NewJSONLWriter() *JSONLWriter {
	return &JSONLWriter{}
}

func (w *JSONLWriter) Write(data interface{}, out io.Writer) error {
	arr, ok := data.([]interface{})
	if !ok {
		buf, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		buf = append(buf, '\n')
		_, err = out.Write(buf)
		return err
	}
	for _, item := range arr {
		buf, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		buf = append(buf, '\n')
		if _, err := out.Write(buf); err != nil {
			return err
		}
	}
	return nil
}

// TextWriter outputs data as a human-readable text summary.
type TextWriter struct{}

// NewTextWriter creates a new text writer.
func NewTextWriter() *TextWriter {
	return &TextWriter{}
}

func (w *TextWriter) Write(data interface{}, out io.Writer) error {
	arr, ok := data.([]interface{})
	if ok {
		for i, item := range arr {
			if i > 0 {
				fmt.Fprintln(out)
			}
			fmt.Fprintf(out, "--- Item %d ---\n", i+1)
			printValue(item, out, 0)
		}
		return nil
	}
	printValue(data, out, 0)
	return nil
}

func printValue(v interface{}, out io.Writer, depth int) {
	indent := strings.Repeat("  ", depth)
	switch val := v.(type) {
	case map[string]interface{}:
		for k, v := range val {
			switch sv := v.(type) {
			case map[string]interface{}:
				fmt.Fprintf(out, "%s%s:\n", indent, k)
				printValue(sv, out, depth+1)
			case []interface{}:
				fmt.Fprintf(out, "%s%s:\n", indent, k)
				printValue(sv, out, depth+1)
			default:
				fmt.Fprintf(out, "%s%s: %v\n", indent, k, v)
			}
		}
	case []interface{}:
		for i, item := range val {
			fmt.Fprintf(out, "%s[%d]:\n", indent, i)
			printValue(item, out, depth+1)
		}
	default:
		fmt.Fprintf(out, "%s%v\n", indent, v)
	}
}

// InfoWriter outputs a summary of JSON structure.
type InfoWriter struct{}

// NewInfoWriter creates a new info writer.
func NewInfoWriter() *InfoWriter {
	return &InfoWriter{}
}

func (w *InfoWriter) Write(data interface{}, out io.Writer) error {
	info := analyzeValue(data)
	fmt.Fprintf(out, "Type: %s\n", info.Type)
	fmt.Fprintf(out, "Depth: %d\n", info.Depth)
	if info.Count > 0 {
		fmt.Fprintf(out, "Count: %d\n", info.Count)
	}
	if len(info.Keys) > 0 {
		fmt.Fprintf(out, "Keys: %s\n", strings.Join(info.Keys, ", "))
	}
	return nil
}

// TypeInfo holds information about a JSON value.
type TypeInfo struct {
	Type  string
	Depth int
	Count int
	Keys  []string
}

func analyzeValue(v interface{}) TypeInfo {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		maxDepth := 0
		for k, child := range val {
			keys = append(keys, k)
			childInfo := analyzeValue(child)
			if childInfo.Depth > maxDepth {
				maxDepth = childInfo.Depth
			}
		}
		return TypeInfo{
			Type:  "object",
			Depth: maxDepth + 1,
			Count: len(val),
			Keys:  keys,
		}
	case []interface{}:
		maxDepth := 0
		for _, item := range val {
			itemInfo := analyzeValue(item)
			if itemInfo.Depth > maxDepth {
				maxDepth = itemInfo.Depth
			}
		}
		return TypeInfo{
			Type:  "array",
			Depth: maxDepth + 1,
			Count: len(val),
		}
	case nil:
		return TypeInfo{Type: "null", Depth: 1}
	case float64:
		return TypeInfo{Type: "number", Depth: 1}
	case string:
		return TypeInfo{Type: "string", Depth: 1}
	case bool:
		return TypeInfo{Type: "boolean", Depth: 1}
	default:
		return TypeInfo{Type: fmt.Sprintf("%T", v), Depth: 1}
	}
}
