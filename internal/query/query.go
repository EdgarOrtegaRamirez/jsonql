// Package query handles JSON reading, path-based querying, sorting, and limiting.
package query

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Reader reads JSON data from various sources.
type Reader struct {
	data interface{}
}

// NewReader creates a Reader from an io.Reader.
func NewReader(r io.Reader) (*Reader, error) {
	var data interface{}
	decoder := json.NewDecoder(r)
	// Read one complete JSON value (handles arrays, objects, etc.)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}
	return &Reader{data: data}, nil
}

// NewReaderFromFile creates a Reader from a file path.
func NewReaderFromFile(path string) (*Reader, error) {
	if path == "-" {
		return NewReader(os.Stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return NewReader(f)
}

// ReadAll reads all JSON from the reader. For JSONL, returns array of values.
func (r *Reader) ReadAll() (interface{}, error) {
	return r.data, nil
}

// Execute parses JSON and returns the whole parsed structure.
func Execute(r *Reader) (interface{}, error) {
	return r.ReadAll()
}

// ExecutePath applies a JSONPath-like query to the data.
// Supported syntax:
//   - "key" → top-level key lookup
//   - "key.subkey" → nested key lookup
//   - "key[*]" → array of objects at key
//   - "key[*].subkey" → extract subkey from array of objects
//   - "[*]" → flatten array elements
//   - "*key" → find all keys containing 'key'
func ExecutePath(r *Reader, path string) (interface{}, error) {
	data, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return applyPath(data, path)
}

// applyPath resolves a path expression against JSON data.
func applyPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := tokenizePath(path)
	return resolve(data, parts)
}

// tokenizePath splits a path into segments, handling [*] as array indices.
func tokenizePath(path string) []string {
	var parts []string
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		if path[i] == '[' && i+1 < len(path) && path[i+1] == '*' {
			// Array index [*]
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			parts = append(parts, "*")
			i += 2 // skip [*]
			if i < len(path) && path[i] == ']' {
				i++ // skip ]
			}
		} else if path[i] == '.' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(path[i])
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// resolve recursively resolves path parts against data.
func resolve(data interface{}, parts []string) (interface{}, error) {
	if len(parts) == 0 {
		return data, nil
	}

	current := parts[0]
	rest := parts[1:]

	// Array index [*]
	if current == "*" {
		arr, ok := data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot index array into non-array value")
		}
		if len(rest) == 0 {
			return arr, nil
		}
		// Recursively resolve rest on each array element
		var results []interface{}
		for _, item := range arr {
			val, err := resolve(item, rest)
			if err != nil {
				continue // skip errors on individual elements
			}
			results = append(results, val)
		}
		return results, nil
	}

	// Object key lookup
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot look up key in non-object")
	}

	val, exists := obj[current]
	if !exists {
		return nil, fmt.Errorf("key %q not found", current)
	}

	if len(rest) == 0 {
		return val, nil
	}
	return resolve(val, rest)
}

// SortResults sorts an array of objects by a field name.
func SortResults(data interface{}, field string, desc bool) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return data, nil // not an array, return as-is
	}

	// Make a copy
	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	sort.SliceStable(sorted, func(i, j int) bool {
		a, _ := getNested(sorted[i], field)
		b, _ := getNested(sorted[j], field)

		ai := toFloat(a)
		bi := toFloat(b)

		if ai != bi {
			return ai < bi
		}

		// Fall back to string comparison
		asc := fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
		return asc
	})

	if desc {
		for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
			sorted[i], sorted[j] = sorted[j], sorted[i]
		}
	}

	return sorted, nil
}

// getNested retrieves a nested value from an object by dot-separated path.
func getNested(data interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	current := data
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[key]
			if !ok {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	return current, true
}

// toFloat converts a JSON value to float64 for comparison.
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

// LimitResults limits the number of items in an array.
func LimitResults(data interface{}, n int) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok || n <= 0 {
		return data, nil
	}
	if n >= len(arr) {
		return data, nil
	}
	return arr[:n], nil
}
