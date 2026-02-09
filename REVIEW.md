# paymo-cli — Project Review

**Date:** 2026-02-09
**Scope:** Architecture, code quality, testing, security, and recommendations

---

## 1. Executive Summary

paymo-cli is a well-structured Go CLI application (~4,700 LOC across 22 files) that provides
terminal-based access to the Paymo time tracking and project management API. The codebase
demonstrates solid engineering fundamentals: clean separation of concerns, good use of Go idioms,
and security-conscious credential handling. However, there are several concrete bugs, missing test
coverage for the command layer, no CI/CD pipeline, and no linting configuration.

**Key Findings:**
- 1 confirmed bug (nil pointer dereference in CSV output)
- 1 code smell (ignored `fmt.Sscanf` errors in rate limit parsing)
- 1 input sanitization gap (unsanitized user input in API `where` clauses)
- 0% test coverage on the `cmd/` package (~1,800 lines untested)
- No CI/CD pipeline or linting configuration
- Good security posture for credential storage and transport

---

## 2. Architecture

### 2.1 Project Structure

```
paymo-cli/
├── main.go                    # Entry point — calls cmd.Execute()
├── cmd/                       # Cobra CLI commands
│   ├── root.go               # Root command, global flags, config init
│   ├── auth.go               # login / logout / status
│   ├── time.go               # start / stop / status / log
│   ├── projects.go           # list / show / create / archive
│   ├── tasks.go              # list / show / create / complete
│   ├── docs.go               # Built-in documentation viewer
│   └── man.go                # Man page / markdown generation
├── internal/
│   ├── api/                  # HTTP client and Paymo API resource layer
│   │   ├── client.go         # HTTP client, auth strategy, rate limiting
│   │   ├── models.go         # Data models (User, Project, Task, TimeEntry)
│   │   ├── projects.go       # Project endpoints
│   │   ├── tasks.go          # Task endpoints
│   │   ├── entries.go        # Time entry endpoints
│   │   └── me.go             # Current user endpoint
│   ├── config/               # Configuration and state persistence
│   │   ├── config.go         # YAML config + JSON credentials
│   │   └── timer.go          # Active timer state (JSON)
│   └── output/               # Output formatting
│       └── output.go         # Table / JSON / CSV formatters
├── go.mod / go.sum
├── README.md
└── SPECIFICATIONS.md
```

### 2.2 Strengths

- **Clean layer separation.** `cmd/` handles CLI concerns, `internal/api/` handles HTTP and
  serialization, `internal/config/` handles persistence, `internal/output/` handles formatting.
  No layer reaches into another's responsibilities.

- **Strategy pattern for authentication.** The `Authenticator` interface with `APIKeyAuth` and
  `BasicAuth` implementations is well-designed and easily extensible.

- **Options pattern for queries.** Typed structs like `ProjectListOptions` and `EntryListOptions`
  provide compile-time safety for filter parameters.

- **Thread-safe rate limiting.** The `sync.Mutex`-guarded rate limit tracking in `client.go` is
  correct and prevents race conditions.

- **Configuration precedence.** The flags → env → file → defaults hierarchy via Viper is standard
  and well-implemented.

### 2.3 Weaknesses

- **No interfaces for testability.** The `cmd/` package directly calls concrete `api.Client`
  methods. Defining an interface for the API layer would enable mocking in command tests without
  HTTP fixtures.

- **No pagination support.** All list endpoints fetch every result in a single call. This will
  break for users with large datasets.

- **`getAPIClient()` is a module-level function, not injected.** This makes the command layer
  hard to test in isolation.

- **`internal/api/models.go` mixes request and response models.** Separating them would make the
  API contract clearer.

---

## 3. Bugs and Code Issues

### 3.1 BUG: Nil pointer dereference in CSV output

**File:** `internal/output/output.go:299`
**Severity:** High — will cause a runtime panic

```go
// Line 290-299
if e.Project != nil {
    projectName = e.Project.Name
}
if e.Task != nil {
    taskName = e.Task.Name
}

w.Write([]string{
    fmt.Sprintf("%d", e.ID),
    fmt.Sprintf("%d", e.Task.ProjectID),  // PANIC if e.Task is nil
    // ...
})
```

`e.Task` is nil-checked above for extracting the name, but then dereferenced unconditionally on
line 299. If a time entry has no associated task, this panics. The fix is to guard the access:

```go
projectID := 0
if e.Task != nil {
    projectID = e.Task.ProjectID
}
// then use projectID in the CSV row
```

### 3.2 CODE SMELL: Ignored errors in `fmt.Sscanf`

**File:** `internal/api/client.go:180-188`
**Severity:** Low — silent failure, rate limiting may not work correctly

```go
fmt.Sscanf(limit, "%d", &c.rateLimit)       // error ignored
fmt.Sscanf(remaining, "%d", &c.rateRemaining) // error ignored
fmt.Sscanf(decay, "%d", &seconds)            // error ignored
```

If the Paymo API returns a non-numeric value in these headers, parsing silently fails and rate
limiting may behave incorrectly. Use `strconv.Atoi` with explicit error handling instead.

### 3.3 INPUT SANITIZATION: Unsanitized user input in `where` clauses

**Files:** `internal/api/projects.go:86`, `internal/api/tasks.go:86`
**Severity:** Medium — depends on Paymo API server-side protections

```go
params.Set("where", fmt.Sprintf("name like \"%%%s%%\"", name))
```

User-supplied `name` is directly interpolated into a query string sent to the Paymo API. A
malicious or accidental input containing `"` or `%` could alter the query semantics. While the
impact depends on how Paymo's API parses the `where` parameter server-side, the client should
escape or validate the input.

---

## 4. Testing

### 4.1 Current State

| Package | Test Files | Approx. LOC | Coverage |
|---------|-----------|-------------|----------|
| `internal/api/` | 3 | ~710 | Partial (client, entries, projects) |
| `internal/config/` | 2 | ~295 | Good (config + timer) |
| `internal/output/` | 1 | ~296 | Good |
| `cmd/` | 0 | 0 | **None** |
| **Total** | **6** | **~1,300** | **~35% of files** |

### 4.2 What's Good

- **Mock HTTP servers.** API tests use `httptest.NewServer()` correctly, avoiding external
  dependencies.
- **Table-driven tests.** `config_test.go` and `timer_test.go` use idiomatic Go table-driven
  patterns.
- **Temp directories.** Tests use `t.TempDir()` for file isolation.
- **Both happy and error paths.** API tests verify successful responses and error conditions.

### 4.3 What's Missing

- **Command-layer tests (`cmd/`).** The entire CLI layer — argument parsing, flag validation,
  output selection, error messaging — is untested. This is the largest gap.
- **Task API tests.** `internal/api/tasks.go` has no dedicated test file (`tasks_test.go` does
  not exist).
- **Edge cases.** No tests for: empty results, unicode in names, very large datasets, concurrent
  timer operations, corrupted config files.
- **Integration / E2E tests.** No tests that exercise the full command → API → output pipeline.
- **Race condition tests.** No `-race` flag usage despite the rate limiter using a mutex.

### 4.4 Recommendations

1. Add `cmd/` tests using Cobra's test utilities (`cmd.SetArgs()`, `cmd.SetOut()`).
2. Add `tasks_test.go` mirroring the existing `projects_test.go` structure.
3. Run tests with `-race` flag to catch data races.
4. Add a CI pipeline that runs `go test -race -cover ./...` on every push.

---

## 5. Security

### 5.1 Positive Practices

| Area | Implementation | Assessment |
|------|---------------|------------|
| Credential storage | `~/.config/paymo-cli/credentials` with `0600` perms | Good |
| Config directory | Created with `0700` permissions | Good |
| Permission validation | `CheckCredentialsPermissions()` warns on loose perms | Good |
| Password handling | Not stored; read via `term.ReadPassword()` | Good |
| Transport | Default `https://` URL | Good |
| No shell execution | Pure HTTP client, no `exec.Command` usage | Good |
| No hardcoded secrets | Test files use dummy values only | Good |

### 5.2 Concerns

1. **Plaintext API key storage.** The API key in `credentials` is not encrypted. Acceptable for
   a CLI tool with proper file permissions, but worth documenting for users.

2. **No HTTPS enforcement.** The `base_url` config value can be overridden to `http://`. The
   client should reject non-HTTPS URLs or at minimum warn.

3. **Where-clause injection.** As noted in section 3.3, user input flows unsanitized into API
   query parameters.

4. **No dependency auditing.** No `govulncheck` or equivalent in the build process.

---

## 6. Code Quality & Developer Experience

### 6.1 Missing Tooling

- **No linting configuration.** No `.golangci-lint.yml` or equivalent. `go vet` is the only
  static analysis available.
- **No CI/CD pipeline.** No `.github/workflows/`, `Makefile`, or equivalent automation.
- **No `Makefile`.** Common Go project convention for `build`, `test`, `lint`, `install` targets.
- **No pre-commit hooks.** No automated checks before commits.

### 6.2 go.mod Issues

- **All dependencies marked `// indirect`.** The direct dependencies (`cobra`, `viper`, `term`)
  should not have the `// indirect` comment. This suggests the module was initialized incorrectly
  or `go mod tidy` was run in an unusual way.

- **Toolchain pinned to `go1.24.13`.** This specific patch version may not be available on all
  systems. Consider using just the minor version (`go 1.24`) or removing the `toolchain` directive.

### 6.3 Code Style

- Consistent formatting (likely `gofmt`-compliant).
- Good use of Go error wrapping with `%w`.
- Descriptive function and variable names.
- Comments on exported functions.
- One inconsistency: some API methods return `(*Type, error)` while others return `(Type, error)`.

---

## 7. Feature Completeness

Based on `SPECIFICATIONS.md` and `README.md`:

| Feature | Status | Notes |
|---------|--------|-------|
| API key auth | Implemented | |
| Email/password auth | Implemented | |
| Time start/stop/status | Implemented | |
| Time log with filters | Implemented | |
| Project CRUD | Implemented | |
| Task CRUD | Implemented | |
| Table output | Implemented | |
| JSON output | Implemented | |
| CSV output | Implemented | Has nil-pointer bug (3.1) |
| Built-in docs | Implemented | |
| Man page generation | Implemented | |
| Rate limiting | Implemented | Has silent parse failures (3.2) |
| Offline timer | Implemented | Via local JSON file |
| Pagination | **Not implemented** | Will fail on large datasets |
| Caching | **Not implemented** | Every command hits the API |
| Shell completions | **Not implemented** | Cobra supports this natively |

---

## 8. Prioritized Recommendations

### P0 — Fix Now

1. **Fix nil pointer dereference** in `internal/output/output.go:299` — this is a crash bug.
2. **Add nil guards** everywhere `e.Task` or `e.Project` are dereferenced in output formatters.

### P1 — Fix Soon

3. **Add `cmd/` test coverage.** This is the largest untested surface area.
4. **Add `tasks_test.go`.** The only API resource without tests.
5. **Sanitize user input** in `where` clause construction (`projects.go:86`, `tasks.go:86`).
6. **Fix `go.mod`** — remove incorrect `// indirect` comments on direct dependencies.
7. **Add CI pipeline** — at minimum: `go vet`, `go test -race ./...`.

### P2 — Improve

8. **Add `golangci-lint` configuration** with sensible defaults.
9. **Add a `Makefile`** with standard targets (`build`, `test`, `lint`, `install`).
10. **Handle `fmt.Sscanf` errors** in rate limit parsing, or switch to `strconv.Atoi`.
11. **Add HTTPS enforcement** — reject or warn on `http://` base URLs.
12. **Add pagination** to list endpoints.
13. **Add shell completions** (`cobra.GenBashCompletion`, etc.).

### P3 — Nice to Have

14. **Add structured logging** behind the `--verbose` flag using `log/slog`.
15. **Define API layer interfaces** for better testability of the command layer.
16. **Add `govulncheck`** to CI for dependency security scanning.
17. **Consider a local cache** (SQLite or similar) to reduce API calls.

---

## 9. Conclusion

paymo-cli is a solid early-stage CLI tool with good architectural foundations. The main areas
needing attention are: one crash bug in CSV output, missing test coverage for the command layer,
and the absence of CI/CD and linting infrastructure. The security posture is appropriate for a
CLI tool of this type. Addressing the P0 and P1 items above would significantly improve the
project's reliability and maintainability.
