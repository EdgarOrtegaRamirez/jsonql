package filter

import (
	"testing"
)

func TestFilterData_BasicEquality(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
		map[string]interface{}{"name": "Charlie", "age": 35},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "age == 30")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
	if arr[0].(map[string]interface{})["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", arr[0])
	}
}

func TestFilterData_GreaterThan(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
		map[string]interface{}{"name": "Charlie", "age": 35},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "age > 28")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}

func TestFilterData_And(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30, "city": "NYC"},
		map[string]interface{}{"name": "Bob", "age": 25, "city": "NYC"},
		map[string]interface{}{"name": "Charlie", "age": 35, "city": "LA"},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "age > 28 and city == \"NYC\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
	if arr[0].(map[string]interface{})["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", arr[0])
	}
}

func TestFilterData_Or(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
		map[string]interface{}{"name": "Charlie", "age": 35},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "age == 25 or age == 35")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}

func TestFilterData_Contains(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice Smith", "age": 30},
		map[string]interface{}{"name": "Bob Jones", "age": 25},
		map[string]interface{}{"name": "Charlie Smith", "age": 35},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "name contains \"Smith\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}

func TestFilterData_StartsWith(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
		map[string]interface{}{"name": "Charlie", "age": 35},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "name startsWith \"A\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
	if arr[0].(map[string]interface{})["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", arr[0])
	}
}

func TestFilterData_Not(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "not age < 28")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
	if arr[0].(map[string]interface{})["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", arr[0])
	}
}

func TestFilterData_EmptyExpression(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice"},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	if len(result.([]interface{})) != len(data) {
		t.Error("expected same data for empty expression")
	}
}

func TestFilterData_NoMatch(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "age > 100")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 0 {
		t.Fatalf("expected 0 results, got %d", len(arr))
	}
}

func TestFilterData_EvalAlias(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.Evaluate(data, "age == 30")
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
}

func TestFilterData_Parentheses(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "(age > 28) and (age < 32)")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
}

func TestFilterData_LessThan(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "age < 28")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
}

func TestFilterData_EndsWith(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "email": "alice@test.com"},
		map[string]interface{}{"name": "Bob", "email": "bob@example.com"},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	result, err := p.FilterData(data, "email endsWith \"test.com\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
}

func TestFilterData_Matches(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "email": "alice@test.com"},
		map[string]interface{}{"name": "Bob", "email": "bob@example.com"},
		map[string]interface{}{"name": "Charlie", "email": "charlie@test.co.uk"},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	// Match emails ending with .com
	result, err := p.FilterData(data, "email matches \".*\\\\.com$\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 results (alice, bob), got %d", len(arr))
	}
}

func TestFilterData_MatchesComplex(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "phone": "555-1234"},
		map[string]interface{}{"name": "Bob", "phone": "555-5678"},
		map[string]interface{}{"name": "Charlie", "phone": "999-0000"},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	// Match phones starting with 555
	result, err := p.FilterData(data, "phone matches \"^555-\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}

func TestFilterData_MatchesInvalidRegex(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"name": "Alice", "email": "alice@test.com"},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	// Invalid regex should not crash, just return empty
	result, err := p.FilterData(data, "email matches \"[invalid\"")
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	arr := result.([]interface{})
	if len(arr) != 0 {
		t.Fatalf("expected 0 results for invalid regex, got %d", len(arr))
	}
}
