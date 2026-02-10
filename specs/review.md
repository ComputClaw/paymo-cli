# Project Review

**Date:** 2026-02-09 (original), updated 2026-02-10
**Scope:** Architecture, code quality, testing, security, and recommendations

---

## 1. Executive Summary

paymo-cli is a well-structured Go CLI application that provides terminal-based access to the
Paymo time tracking and project management API. The codebase demonstrates solid engineering
fundamentals: clean separation of concerns, good use of Go idioms, and security-conscious
credential handling.

**Status of findings from original review:**
- Nil pointer dereference in CSV output — **fixed**
- 0% cmd/ test coverage — **fixed** (full mock-based test suite added)
- No CI/CD pipeline — **fixed** (CI + GoReleaser release pipeline added)
- Missing time edit/delete commands — **fixed** (show/edit/delete added)
- Missing `--quiet` flag — **fixed**
- Missing `--no-cache` flag — **fixed** (BoltDB cache layer added)
- Missing JSON output for mutation commands — **fixed**
- Missing structured error output — **fixed**
- Missing machine-readable command discovery — **fixed** (`paymo schema`)

**Remaining findings:**
- `fmt.Sscanf` ignored errors in rate limit parsing (low severity)
- Unsanitized user input in API `where` clauses (medium severity)
- No pagination support for list endpoints

---

## 2. Spec Conformance

### 2.1 Implemented (Matching or Close to Spec)
- API key and email/password authentication
- Time start/stop/status/log/show/edit/delete
- Project list/show/create/archive
- Task list/show/create/complete
- Client listing
- Table, JSON, CSV output formats
- `--quiet` flag for minimal output
- `--no-cache` flag to bypass cache
- BoltDB cache with stale fallback (spec said SQLite — BoltDB is simpler/better fit)
- CI pipeline with `go test -race -cover`
- GitHub Actions release pipeline via GoReleaser
- Shell completions (via Cobra)
- Machine-readable schema (`paymo schema`)

### 2.2 Deviations from Spec
| Area | Spec Says | Implementation Does |
|------|-----------|-------------------|
| Config location | `~/.paymo.yaml` | `~/.config/paymo-cli/` directory |
| Cache backend | SQLite | BoltDB |
| JSON output | Wrapped: `{"time_entries": [...], "total_count": N}` | Raw arrays `[{...}]` |
| Auth interface | `AuthMethod` with `Validate()` | `Authenticator` without `Validate()` |

### 2.3 Not Yet Implemented
- `paymo projects update` — cannot update project fields
- `paymo tasks update` — cannot update task fields
- `paymo auth refresh` — no session token refresh
- Session token authentication
- `paymo reports` command group
- Pagination on list endpoints
- `--sort` flag on lists
- `--user` filter flag

---

## 3. AI-Friendliness Assessment

### 3.1 What Works Well
- `paymo schema` provides machine-readable command discovery
- All commands support `--format json` including mutations
- Structured JSON error output when `--format json`
- `--quiet` flag returns only IDs for mutations
- Consistent ID-or-name resolution on project/task commands
- Comprehensive `--help` with examples on all commands
- AI agent guide published at https://mbundgaard.github.io/paymo-cli/

### 3.2 Remaining Gaps
- JSON output uses bare arrays instead of metadata envelope
- Single exit code (1) for all error types — no distinction between auth/not-found/rate-limit
- `--verbose` flag defined but barely used

### 3.3 Recommended Exit Code Scheme
- `0` — success
- `1` — general error
- `2` — usage/argument error
- `3` — authentication error
- `4` — not found
- `5` — rate limited
- `6` — network/API error

---

## 4. Bugs and Code Issues

### 4.1 Ignored errors in `fmt.Sscanf` (Low)
**File:** `internal/api/client.go`

```go
fmt.Sscanf(limit, "%d", &c.rateLimit)       // error ignored
fmt.Sscanf(remaining, "%d", &c.rateRemaining) // error ignored
fmt.Sscanf(decay, "%d", &seconds)            // error ignored
```

Non-numeric rate limit headers silently fail. Use `strconv.Atoi` with error handling.

### 4.2 Unsanitized user input in `where` clauses (Medium)
**Files:** `internal/api/projects.go`, `internal/api/tasks.go`

```go
params.Set("where", fmt.Sprintf("name like \"%%%s%%\"", name))
```

User-supplied names are interpolated directly into query strings. Input containing `"` or `%` could alter query semantics.

---

## 5. Security

### 5.1 Positive Practices
- Credential file stored with `0600` permissions
- Config directory created with `0700` permissions
- `CheckCredentialsPermissions()` warns on loose perms
- Passwords read via `term.ReadPassword()`, never stored
- HTTPS by default
- No shell execution (`exec.Command`) anywhere
- No hardcoded secrets

### 5.2 Concerns
- API key stored in plaintext (acceptable with file permissions, but worth documenting)
- `base_url` can be overridden to `http://` — no HTTPS enforcement
- Where-clause injection (see 4.2)
- No `govulncheck` in CI

---

## 6. Code Quality

### Strengths
- Clean layer separation (cmd → api → config → output)
- Strategy pattern for authentication
- Options pattern for query parameters
- Thread-safe rate limiting
- Consistent formatting (`gofmt`-compliant)
- Good error wrapping with `%w`

### Gaps
- No linting configuration (`.golangci-lint.yml`)
- No `Makefile`
- No pre-commit hooks
- No pagination support
