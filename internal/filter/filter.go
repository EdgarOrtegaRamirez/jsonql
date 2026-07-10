// Package filter provides a simple expression parser for filtering JSON data.
// Supports: ==, !=, >, >=, <, <=, &&, ||, and, or, not, contains, startsWith, endsWith
package filter

import (
	"fmt"
	"sort"
	"strings"
)

// Parser holds filter expression parsing state.
type Parser struct{}

// NewParser creates a new filter parser.
func NewParser() (*Parser, error) {
	return &Parser{}, nil
}

// FilterData filters an array of objects based on a filter expression.
func (p *Parser) FilterData(data interface{}, expr string) (interface{}, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return data, nil
	}

	tokens, err := tokenize(trimmed)
	if err != nil {
		return nil, fmt.Errorf("tokenize: %w", err)
	}

	ast, err := parseExpr(tokens)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	arr, ok := data.([]interface{})
	if !ok {
		match, err := matchSingle(ast, data)
		if err != nil {
			return nil, err
		}
		if match {
			return []interface{}{data}, nil
		}
		return []interface{}{}, nil
	}

	var results []interface{}
	for _, item := range arr {
		match, err := matchSingle(ast, item)
		if err != nil {
			continue
		}
		if match {
			results = append(results, item)
		}
	}
	if results == nil {
		results = []interface{}{}
	}
	return results, nil
}

// Evaluate is an alias for FilterData for compatibility.
func (p *Parser) Evaluate(data interface{}, expr string) (interface{}, error) {
	return p.FilterData(data, expr)
}

// --- Token types ---

type tokenType int

const (
	tokenString tokenType = iota
	tokenNumber
	tokenField
	tokenOp
	tokenBool
	tokenAnd
	tokenOr
	tokenNot
)

type token struct {
	typ   tokenType
	value string
}

// --- Tokenizer ---

func tokenize(expr string) ([]token, error) {
	var tokens []token

	for len(expr) > 0 {
		expr = skipWS(expr)
		if expr == "" {
			break
		}

		// Two-char operators
		if len(expr) >= 2 {
			switch expr[:2] {
			case "==", "!=", ">=", "<=":
				tokens = append(tokens, token{tokenOp, expr[:2]})
				expr = expr[2:]
				continue
			case "&&":
				tokens = append(tokens, token{tokenAnd, "&&"})
				expr = expr[2:]
				continue
			case "||":
				tokens = append(tokens, token{tokenOr, "||"})
				expr = expr[2:]
				continue
			}
		}

		// Single-char operators
		switch expr[0] {
		case '>', '<':
			tokens = append(tokens, token{tokenOp, expr[:1]})
			expr = expr[1:]
			continue
		case '(':
			tokens = append(tokens, token{tokenOp, "("})
			expr = expr[1:]
			continue
		case ')':
			tokens = append(tokens, token{tokenOp, ")"})
			expr = expr[1:]
			continue
		}

		// Keywords
		for _, kw := range []struct {
			name string
			tt   tokenType
		}{
			{"and", tokenAnd}, {"or", tokenOr}, {"not", tokenNot},
			{"contains", tokenOp}, {"startsWith", tokenOp}, {"endsWith", tokenOp},
		} {
			prefix := kw.name + " "
			if expr == kw.name || strings.HasPrefix(expr, prefix) {
				tokens = append(tokens, token{kw.tt, kw.name})
				expr = expr[len(kw.name):]
				expr = skipWS(expr)
				goto next
			}
		}

		// Strings
		if expr[0] == '"' || expr[0] == '\'' {
			q := expr[0]
			expr = expr[1:]
			var sb strings.Builder
			for len(expr) > 0 && expr[0] != q {
				if expr[0] == '\\' && len(expr) > 1 {
					sb.WriteByte(expr[1])
					expr = expr[2:]
				} else {
					sb.WriteByte(expr[0])
					expr = expr[1:]
				}
			}
			if len(expr) == 0 {
				return nil, fmt.Errorf("unterminated string")
			}
			tokens = append(tokens, token{tokenString, sb.String()})
			expr = expr[1:]
			continue
		}

		// Numbers
		if isDigit(expr[0]) || (expr[0] == '-' && len(expr) > 1 && isDigit(expr[1])) {
			var sb strings.Builder
			if expr[0] == '-' {
				sb.WriteByte('-')
				expr = expr[1:]
			}
			for len(expr) > 0 && (isDigit(expr[0]) || expr[0] == '.') {
				sb.WriteByte(expr[0])
				expr = expr[1:]
			}
			tokens = append(tokens, token{tokenNumber, sb.String()})
			continue
		}

		// Identifiers
		if isIdentStart(expr[0]) {
			var sb strings.Builder
			for len(expr) > 0 && isIdentChar(expr[0]) {
				sb.WriteByte(expr[0])
				expr = expr[1:]
			}
			val := sb.String()
			tt := tokenField
			if strings.EqualFold(val, "true") {
				tt = tokenBool
			} else if strings.EqualFold(val, "false") {
				tt = tokenBool
			}
			tokens = append(tokens, token{tt, val})
			continue
		}

		return nil, fmt.Errorf("unexpected character %q", expr[0])
	next:
	}

	return tokens, nil
}

func skipWS(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '.'
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || isDigit(c) || c == '-' || c == '/'
}

// --- AST ---

type nodeType int

const (
	nodeLiteral nodeType = iota
	nodeField
	nodeOp
	nodeUnary
	nodeStrMethod
)

type node struct {
	typ    nodeType
	op     string
	left   *node
	right  *node
	value  interface{}
	method string
	field  string
}

// parseAtom with index-based approach
type parserState struct {
	tokens []token
	idx    int
}

func newParserState(tokens []token) *parserState {
	return &parserState{tokens: tokens, idx: 0}
}

func (ps *parserState) peek() (*token, bool) {
	if ps.idx >= len(ps.tokens) {
		return nil, false
	}
	return &ps.tokens[ps.idx], true
}

func (ps *parserState) advance() {
	ps.idx++
}

func (ps *parserState) remaining() int {
	return len(ps.tokens) - ps.idx
}

func parseExpr(tokens []token) (*node, error) {
	ps := newParserState(tokens)
	result, err := parseOrState(ps)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseOrState(ps *parserState) (*node, error) {
	left, err := parseAndState(ps)
	if err != nil {
		return nil, err
	}
	for {
		tok, ok := ps.peek()
		if !ok || tok.typ != tokenOr {
			break
		}
		ps.advance()
		right, err := parseAndState(ps)
		if err != nil {
			return nil, err
		}
		left = &node{typ: nodeOp, op: "or", left: left, right: right}
	}
	return left, nil
}

func parseAndState(ps *parserState) (*node, error) {
	left, err := parseNotState(ps)
	if err != nil {
		return nil, err
	}
	for {
		tok, ok := ps.peek()
		if !ok || tok.typ != tokenAnd {
			break
		}
		ps.advance()
		right, err := parseNotState(ps)
		if err != nil {
			return nil, err
		}
		left = &node{typ: nodeOp, op: "and", left: left, right: right}
	}
	return left, nil
}

func parseNotState(ps *parserState) (*node, error) {
	tok, ok := ps.peek()
	if !ok {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	if tok.typ == tokenNot {
		ps.advance()
		sub, err := parseNotState(ps)
		if err != nil {
			return nil, err
		}
		return &node{typ: nodeUnary, op: "not", left: sub}, nil
	}
	return parseCompareState(ps)
}

func parseCompareState(ps *parserState) (*node, error) {
	if ps.remaining() == 0 {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	left, err := parseAtomState(ps)
	if err != nil {
		return nil, err
	}
	tok, ok := ps.peek()
	if ok && tok.typ == tokenOp {
		op := tok.value
		if op == "(" || op == ")" {
			return left, nil
		}
		ps.advance()
		right, err := parseAtomState(ps)
		if err != nil {
			return nil, err
		}
		return &node{typ: nodeOp, op: op, left: left, right: right}, nil
	}
	return left, nil
}

func parseAtomState(ps *parserState) (*node, error) {
	if ps.remaining() == 0 {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	tok := ps.tokens[ps.idx]

	// Parenthesized: find matching )
	if tok.value == "(" {
		ps.advance() // skip (
		body, err := parseOrState(ps)
		if err != nil {
			return nil, err
		}
		// Skip closing paren
		if closeTok, ok := ps.peek(); ok && closeTok.value == ")" {
			ps.advance()
		}
		return body, nil
	}

	// String with method: field "contains" value OR field contains "value"
	if tok.typ == tokenField && ps.remaining() >= 2 {
		nextTok := ps.tokens[ps.idx+1]
		if nextTok.typ == tokenString && ps.remaining() >= 3 && ps.tokens[ps.idx+2].typ == tokenString {
			// Pattern: field "contains" "value"
			method := nextTok.value
			if method == "contains" || method == "startsWith" || method == "endsWith" {
				strVal := ps.tokens[ps.idx+2].value
				ps.advance()
				ps.advance()
				ps.advance()
				return &node{
					typ:    nodeStrMethod,
					method: method,
					field:  tok.value,
					value:  strVal,
				}, nil
			}
		}
		if nextTok.typ == tokenOp && nextTok.value == "contains" && ps.remaining() >= 3 && ps.tokens[ps.idx+2].typ == tokenString {
			// Pattern: field contains "value"
			strVal := ps.tokens[ps.idx+2].value
			ps.advance()
			ps.advance()
			ps.advance()
			return &node{
				typ:    nodeStrMethod,
				method: "contains",
				field:  tok.value,
				value:  strVal,
			}, nil
		}
		if nextTok.typ == tokenOp && nextTok.value == "startsWith" && ps.remaining() >= 3 && ps.tokens[ps.idx+2].typ == tokenString {
			strVal := ps.tokens[ps.idx+2].value
			ps.advance()
			ps.advance()
			ps.advance()
			return &node{
				typ:    nodeStrMethod,
				method: "startsWith",
				field:  tok.value,
				value:  strVal,
			}, nil
		}
		if nextTok.typ == tokenOp && nextTok.value == "endsWith" && ps.remaining() >= 3 && ps.tokens[ps.idx+2].typ == tokenString {
			strVal := ps.tokens[ps.idx+2].value
			ps.advance()
			ps.advance()
			ps.advance()
			return &node{
				typ:    nodeStrMethod,
				method: "endsWith",
				field:  tok.value,
				value:  strVal,
			}, nil
		}
	}
	_ = tok // re-read after advance checks
	tok = ps.tokens[ps.idx]

	// String literal
	if tok.typ == tokenString {
		ps.advance()
		return &node{typ: nodeLiteral, value: tok.value}, nil
	}

	// Number
	if tok.typ == tokenNumber {
		ps.advance()
		return &node{typ: nodeLiteral, value: tok.value}, nil
	}

	// Bool
	if tok.typ == tokenBool {
		ps.advance()
		val := tok.value == "true"
		return &node{typ: nodeLiteral, value: val}, nil
	}

	// Field
	if tok.typ == tokenField {
		ps.advance()
		return &node{typ: nodeField, field: tok.value}, nil
	}

	ps.advance()
	return nil, fmt.Errorf("unexpected token %q", tok.value)
}

// --- Evaluator ---

func eval(n *node, data interface{}) (interface{}, error) {
	switch n.typ {
	case nodeOp:
		l, err := eval(n.left, data)
		if err != nil {
			return nil, err
		}
		r, err := eval(n.right, data)
		if err != nil {
			return nil, err
		}
		return evalBinary(n.op, l, r)
	case nodeUnary:
		v, err := eval(n.left, data)
		if err != nil {
			return nil, err
		}
		if n.op == "not" {
			return !isTruthy(v), nil
		}
		return v, nil
	case nodeLiteral:
		return n.value, nil
	case nodeField:
		if arr, ok := data.([]interface{}); ok {
			// For comparison: return first matching field value
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					if v, exists := m[n.field]; exists {
						return v, nil
					}
				}
			}
			return nil, fmt.Errorf("field %q not found", n.field)
		}
		if m, ok := data.(map[string]interface{}); ok {
			v, exists := m[n.field]
			if !exists {
				return nil, fmt.Errorf("field %q not found", n.field)
			}
			return v, nil
		}
		return nil, fmt.Errorf("cannot access field in non-object")
	}
	return nil, fmt.Errorf("unknown node type: %d", n.typ)
}

func evalBinary(op string, l, r interface{}) (interface{}, error) {
	if op == "or" || op == "||" {
		return isTruthy(l) || isTruthy(r), nil
	}
	if op == "and" || op == "&&" {
		return isTruthy(l) && isTruthy(r), nil
	}
	return compare(op, l, r)
}

func compare(op string, l, r interface{}) (interface{}, error) {
	// Check types
	_, lIsStr := l.(string)
	_, rIsStr := r.(string)

	if !lIsStr && !rIsStr {
		// Both non-string: use numeric
		return compareNumeric(op, l, r)
	}
	if lIsStr && rIsStr {
		// Both strings: use string comparison
		return compareString(op, l, r)
	}
	// Mixed: try numeric (e.g., float64 value vs string literal "100")
	return compareNumeric(op, l, r)
}

func compareNumeric(op string, l, r interface{}) (interface{}, error) {
	la := toFloat(l)
	rb := toFloat(r)
	var cmp float64
	switch op {
	case "==":
		cmp = boolToInt(la == rb)
	case "!=":
		cmp = boolToInt(la != rb)
	case ">":
		cmp = boolToInt(la > rb)
	case ">=":
		cmp = boolToInt(la >= rb)
	case "<":
		cmp = boolToInt(la < rb)
	case "<=":
		cmp = boolToInt(la <= rb)
	default:
		return nil, fmt.Errorf("unknown operator: %s", op)
	}
	return cmp, nil
}

func compareString(op string, l, r interface{}) (interface{}, error) {
	lv := fmt.Sprintf("%v", l)
	rv := fmt.Sprintf("%v", r)
	var cmp float64
	switch op {
	case "==":
		cmp = boolToInt(lv == rv)
	case "!=":
		cmp = boolToInt(lv != rv)
	case ">":
		cmp = boolToInt(lv > rv)
	case ">=":
		cmp = boolToInt(lv >= rv)
	case "<":
		cmp = boolToInt(lv < rv)
	case "<=":
		cmp = boolToInt(lv <= rv)
	default:
		return nil, fmt.Errorf("unknown operator: %s", op)
	}
	return cmp, nil
}

func boolToInt(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func isTruthy(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	case nil:
		return false
	default:
		return true
	}
}

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

// --- Array matching ---

func matchSingle(ast *node, item interface{}) (bool, error) {
	switch ast.typ {
	case nodeOp:
		l, err := matchSingle(ast.left, item)
		if err != nil {
			return false, err
		}
		r, err := matchSingle(ast.right, item)
		if err != nil {
			return false, err
		}
		if ast.op == "and" || ast.op == "&&" {
			return l && r, nil
		}
		if ast.op == "or" || ast.op == "||" {
			return l || r, nil
		}
		lv, err := eval(ast.left, item)
		if err != nil {
			return false, err
		}
		rv, err := eval(ast.right, item)
		if err != nil {
			return false, err
		}
		res, err := compare(ast.op, lv, rv)
		if err != nil {
			return false, err
		}
		return res == float64(1), nil

	case nodeUnary:
		val, err := matchSingle(ast.left, item)
		if err != nil {
			return false, err
		}
		if ast.op == "not" {
			return !val, nil
		}
		return val, nil

	case nodeStrMethod:
		if m, ok := item.(map[string]interface{}); ok {
			return matchStrMethod(m, ast.field, ast.method, ast.value.(string)), nil
		}
		return false, nil

	case nodeLiteral:
		return true, nil

	case nodeField:
		if m, ok := item.(map[string]interface{}); ok {
			_, ok := m[ast.field]
			return ok, nil
		}
		return false, nil
	}
	return false, nil
}

func matchStrMethod(obj map[string]interface{}, field, method, value string) bool {
	val, exists := obj[field]
	if !exists {
		return false
	}
	str := fmt.Sprintf("%v", val)
	switch method {
	case "contains":
		return strings.Contains(str, value)
	case "startsWith":
		return strings.HasPrefix(str, value)
	case "endsWith":
		return strings.HasSuffix(str, value)
	}
	return false
}

// SortStringResults sorts an array of items by string representation.
func SortStringResults(items []interface{}, desc bool) []interface{} {
	sort.SliceStable(items, func(i, j int) bool {
		a := fmt.Sprintf("%v", items[i])
		b := fmt.Sprintf("%v", items[j])
		return a < b
	})
	if desc {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}
	return items
}
