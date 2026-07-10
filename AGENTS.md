# AGENTS.md — for AI Coding Agents

## Project Overview

**jsonql** is a lightweight JSON query CLI with simple path syntax, filtering, sorting, and multiple output formats. Written in Go with zero runtime dependencies.

## Key Files

| File | Purpose |
|------|---------|
| `cmd/jsonql/main.go` | CLI entry point with Cobra commands |
| `internal/query/query.go` | JSON reading, path resolution, sort, limit |
| `internal/filter/filter.go` | Expression tokenizer, recursive descent parser, evaluator |
| `internal/output/output.go` | Formatters: JSON, JSONL, text, info |

## Architecture

```
CLI (Cobra)
  ├── root command (help, version)
  ├── query command → Reader → path/query → filter → sort → limit → output
  ├── pretty command → Reader → JSON output
  └── info command → Reader → structure analysis → info output
```

## Build & Test

```bash
go build ./cmd/jsonql/
go test ./... -count=1
go vet ./...
```

## Filter Expression Language

The filter parser uses a recursive descent parser with these precedence levels (lowest to highest):

1. **OR** (`or`, `||`)
2. **AND** (`and`, `&&`)
3. **NOT** (`not`)
4. **Comparison** (`==`, `!=`, `>`, `<`, `>=`, `<=`)
5. **Atoms**: literals (strings, numbers, booleans), field references, parenthesized expressions
6. **String methods**: `field "contains" value`, `field "startsWith" value`, `field "endsWith" value`

### Key Implementation Details

- **Tokenizer** handles keywords with space boundaries to avoid matching `contains` as part of identifiers
- **String method detection** checks if a field token is followed by a string token that is one of the method names
- **Evaluator** uses `toFloat()` for numeric comparison, falling back to string comparison when values don't convert

## Testing

```bash
# Run all tests
go test ./... -count=1

# Run with verbose output
go test ./... -count=1 -v

# Run with race detector
go test -race ./...
```

## Adding New Output Formats

1. Add a new `Format` constant in `internal/output/output.go`
2. Create a new writer struct implementing the `Writer` interface
3. Update the switch in `main.go` to handle the new format
4. Add tests in `internal/output/output_test.go`

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- Standard library only for JSON processing
