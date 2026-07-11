# jsonql

Ergonomic JSON query CLI — simple path syntax, filtering, sorting, and multiple output formats.

## What It Does

`jsonql` makes it easy to extract, filter, sort, and format data from JSON files directly from the command line. Unlike `jq` which has a steep learning curve, `jsonql` uses intuitive path syntax and natural-language filter expressions.

## Features

- **Simple path queries** — `users[*].name` extracts names from an array of objects
- **Filter expressions** — `age > 25 and city == "NYC"` with natural `and`/`or`/`not` keywords
- **String methods** — `name contains "Smith"`, `email startsWith "alice"`, `path endsWith ".json"`, `email matches ".*\\.com$"` with regex support
- **Sorting** — `--sort age --sort-desc` for descending order
- **Limiting** — `--limit 10` to cap results
- **Multiple output formats** — JSON (pretty/compact), JSONL, human-readable text
- **Structure inspection** — `jsonql info` shows types, keys, counts, depth
- **Pretty printing** — `jsonql pretty` formats JSON with indentation
- **Zero dependencies** — Single binary, no runtime dependencies
- **Stdin support** — Pipe JSON from other tools: `curl ... | jsonql --path data`

## Installation

```bash
# From source
go install github.com/EdgarOrtegaRamirez/jsonql@latest

# Or build locally
git clone https://github.com/EdgarOrtegaRamirez/jsonql.git
cd jsonql
go build -o jsonql ./cmd/jsonql/
```

## Quick Start

```bash
# Pretty-print JSON
jsonql pretty data.json

# Show structure summary
jsonql info data.json

# Extract a top-level key
jsonql --path name data.json

# Extract from nested structure
jsonql --path 'users[*].name' data.json

# Filter by condition
jsonql --path 'users[*]' --filter 'age > 25' data.json

# Sort and limit
jsonql --path 'items[*]' --sort price --limit 5 data.json

# Output as JSONL
jsonql --path 'items[*]' --format jsonl data.json

# Pipe from stdin
cat data.json | jsonql --path users

# Complex filter with string methods
jsonql --path 'logs[*]' --filter 'level contains "ERROR"' data.json

# Regex filter (match emails ending in .com)
jsonql --path 'users[*]' --filter 'email matches ".*\\.com$"' data.json

# Regex filter (match phone numbers starting with 555)
jsonql --path 'contacts[*]' --filter 'phone matches "^555-"' data.json
```

## Commands

### `query` (default)

Query JSON files with path, filter, sort, and limit options.

```
jsonql --path <path> --filter <expr> --sort <field> --limit <n> --format <fmt> [file...]
```

### `pretty`

Pretty-print JSON from files or stdin.

```
jsonql pretty [--compact] [file...]
```

### `info`

Show JSON structure summary (types, keys, counts, depth).

```
jsonql info [file...]
```

## Path Syntax

| Syntax | Description |
|--------|-------------|
| `key` | Top-level key lookup |
| `key.subkey` | Nested key lookup |
| `key[*]` | All elements in array at key |
| `key[*].subkey` | Extract subkey from each array element |
| `[*]` | Flatten array elements |

## Filter Expressions

| Syntax | Description |
|--------|-------------|
| `field == value` | Equality (numbers, strings) |
| `field != value` | Inequality |
| `field > value` | Greater than |
| `field < value` | Less than |
| `field >= value` | Greater or equal |
| `field <= value` | Less or equal |
| `field contains "str"` | String contains substring |
| `field startsWith "str"` | String starts with prefix |
| `field endsWith "str"` | String ends with suffix |
| `field matches "regex"` | String matches regex pattern |
| `expr and expr` | Logical AND |
| `expr or expr` | Logical OR |
| `not expr` | Logical NOT |
| `(expr)` | Grouping |

## Output Formats

| Format | Description |
|--------|-------------|
| `json` | Formatted JSON (default) |
| `jsonl` | One JSON object per line |
| `text` | Human-readable text with indentation |

## Architecture

```
CLI (Cobra)
  ├── query command → Reader → Query → Filter → Sort → Limit → Output
  ├── pretty command → Reader → JSON output
  └── info command → Reader → Structure analysis → Info output
```

- **query/query.go**: JSON reading, path resolution, sorting, limiting
- **filter/filter.go**: Expression tokenizer, recursive descent parser, evaluator
- **output/output.go**: JSON, JSONL, text, and info formatters

## Testing

```bash
go test ./... -count=1
go test ./... -count=1 -race
```

## License

MIT — see [LICENSE](LICENSE) for details.
