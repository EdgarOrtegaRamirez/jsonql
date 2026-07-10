# Security

## Security Considerations

jsonql is a local CLI tool with no network calls. All processing happens on local files or stdin.

### Input Validation

- File paths are opened directly; no path traversal since we don't use shell execution
- JSON input is parsed using Go's standard `encoding/json` which is safe against malformed input
- Filter expressions are parsed by a recursive descent parser — no `eval()` or code execution

### File Safety

- Files are opened with `os.Open()` — standard safe file operations
- No temporary files are created
- No file permissions are modified

### Dependencies

All dependencies are pinned to specific versions in `go.mod`:
- `github.com/spf13/cobra v1.10.2` — well-maintained CLI framework

### No Secrets

- No hardcoded secrets, tokens, or API keys
- No network calls
- No environment variable reading

## Reporting

If you discover a security vulnerability, please open a GitHub issue.
