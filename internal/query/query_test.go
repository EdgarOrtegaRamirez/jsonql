package query

import (
	"strings"
	"testing"
)

func TestExecutePath_TopLevelKey(t *testing.T) {
	jsonStr := `{"name": "Alice", "age": 30}`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	result, err := ExecutePath(reader, "name")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result != "Alice" {
		t.Errorf("expected Alice, got %v", result)
	}
}

func TestExecutePath_NestedKey(t *testing.T) {
	jsonStr := `{"user": {"name": "Bob", "email": "bob@example.com"}}`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	result, err := ExecutePath(reader, "user.name")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result != "Bob" {
		t.Errorf("expected Bob, got %v", result)
	}
}

func TestExecutePath_ArrayIndex(t *testing.T) {
	jsonStr := `{"users": [{"name": "Alice"}, {"name": "Bob"}]}`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	result, err := ExecutePath(reader, "users[*].name")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr))
	}
	if arr[0] != "Alice" || arr[1] != "Bob" {
		t.Errorf("expected [Alice, Bob], got %v", arr)
	}
}

func TestExecutePath_FlatArray(t *testing.T) {
	jsonStr := `[1, 2, 3, 4, 5]`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	result, err := ExecutePath(reader, "[*]")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(arr))
	}
}

func TestExecutePath_NonExistentKey(t *testing.T) {
	jsonStr := `{"name": "Alice"}`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	_, err = ExecutePath(reader, "missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestSortResults(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Charlie", "age": 30},
		map[string]interface{}{"name": "Alice", "age": 25},
		map[string]interface{}{"name": "Bob", "age": 35},
	}

	result, err := SortResults(data, "age", false)
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	arr := result.([]interface{})
	if arr[0].(map[string]interface{})["name"] != "Alice" {
		t.Errorf("expected Alice first, got %v", arr[0])
	}
	if arr[2].(map[string]interface{})["name"] != "Bob" {
		t.Errorf("expected Bob last, got %v", arr[2])
	}
}

func TestLimitResults(t *testing.T) {
	data := []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}
	result, err := LimitResults(data, 3)
	if err != nil {
		t.Fatalf("limit: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 3 {
		t.Fatalf("expected 3, got %d", len(arr))
	}
	if arr[2] != float64(3) {
		t.Errorf("expected 3 at index 2, got %v", arr[2])
	}
}

func TestLimitResults_NoLimit(t *testing.T) {
	data := []interface{}{1, 2, 3}
	result, err := LimitResults(data, 0)
	if err != nil {
		t.Fatalf("limit: %v", err)
	}
	if len(result.([]interface{})) != len(data) {
		t.Error("expected same data when limit is 0")
	}
}

func TestExecutePath_MissingKey(t *testing.T) {
	jsonStr := `{"a": 1}`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	_, err = ExecutePath(reader, "b")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestExecutePath_EmptyPath(t *testing.T) {
	jsonStr := `{"name": "Alice"}`
	reader, err := NewReader(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}

	result, err := ExecutePath(reader, "")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	_, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("expected map, got %T", result)
	}
}

func TestSortResults_Desc(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 25},
		map[string]interface{}{"name": "Bob", "age": 35},
		map[string]interface{}{"name": "Charlie", "age": 30},
	}

	result, err := SortResults(data, "age", true)
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	arr := result.([]interface{})
	if arr[0].(map[string]interface{})["name"] != "Bob" {
		t.Errorf("expected Bob first (desc), got %v", arr[0])
	}
}

func TestLimitResults_ExceedsLength(t *testing.T) {
	data := []interface{}{1, 2}
	result, err := LimitResults(data, 10)
	if err != nil {
		t.Fatalf("limit: %v", err)
	}
	if len(result.([]interface{})) != len(data) {
		t.Error("expected same data when limit exceeds length")
	}
}
