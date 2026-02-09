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

## 2. Conformance to SPECIFICATIONS.md

The project includes a detailed `SPECIFICATIONS.md` that serves as the design blueprint. This
section tracks what was specified versus what was actually built.

### 2.1 Specified but NOT Implemented

| Specified Feature | Status | Impact |
|-------------------|--------|--------|
| **SQLite caching layer** (`internal/cache/`) | Not built | Every command hits the API; no offline mode |
| **Offline mode with sync** | Not built | `paymo time sync` does not exist |
| **`paymo reports` command group** | Not built | `reports time`, `reports projects`, `reports export` all missing |
| **`paymo time edit <id>`** | Not built | Cannot edit time entries from CLI |
| **`paymo time delete <id>`** | Not built | Cannot delete time entries from CLI |
| **`paymo projects update`** | Not built | Cannot update project fields |
| **`paymo tasks update`** | Not built | Cannot update task fields |
| **`paymo auth refresh`** | Not built | No session token refresh |
| **Session token authentication** | Not built | Only API key and basic auth |
| **Shell completions** | Not built | Cobra supports this natively |
| **`--no-cache` global flag** | Not built | No cache exists to bypass |
| **`--sort/-s` flag on lists** | Not built | No sorting support |
| **`--user` filter flag** | Not built | Cannot filter by user |
| **`pkg/models/` shared models** | Not built | Models live in `internal/api/models.go` |
| **`scripts/` directory** | Not built | No build/release scripts |
| **`docs/` directory** | Not built | Docs are embedded in `cmd/docs.go` |
| **GitHub Actions release pipeline** | Not built | No CI/CD at all |
| **Homebrew formula** | Not built | No distribution packaging |
| **Integration and E2E tests** | Not built | Only unit tests exist |

### 2.2 Implemented but Deviates from Spec

| Area | Spec Says | Implementation Does |
|------|-----------|-------------------|
| **Config location** | `~/.paymo.yaml` (single file) | `~/.config/paymo-cli/` (directory with 3 files) |
| **Cache location** | `~/.cache/paymo` with SQLite | No cache — timer state in `~/.config/paymo-cli/timer.json` |
| **JSON output format** | Wrapped: `{"time_entries": [...], "total_duration": N}` | Raw arrays: `[{...}, {...}]` — no metadata envelope |
| **Project structure** | `internal/api/auth.go` separate auth file | Auth is part of `internal/api/client.go` |
| **Project structure** | Separate `table.go`, `json.go`, `csv.go` | Single `output.go` with all formatters |
| **Auth method** | `AuthMethod` interface with `Validate()` | `Authenticator` interface without `Validate()` |
| **Hashed credentials** | "Store hashed credentials securely" | Plaintext JSON with file permissions |

### 2.3 Assessment

Roughly **40-50% of the specification has been implemented**. The core time tracking, project
management, and task management commands work. However, the caching/offline layer, reports
module, CRUD update/delete operations, and distribution infrastructure are entirely missing.

The JSON output format deviation is particularly important: the spec calls for a metadata
envelope (`{"time_entries": [...], "total_duration": N, "total_count": N}`) but the
implementation outputs raw arrays. This matters for AI consumers that need to know result
counts without iterating.

---

## 3. AI-Friendliness Assessment

The README states this tool is "AI-Friendly: Consistent output formats and comprehensive
`--help` for agent use." This section evaluates that claim.

### 3.1 Discoverability — MODERATE

**What works:**
- `paymo --help` shows all top-level command groups with descriptions
- Each subcommand has `Short`, `Long`, and `Example` fields in help text
- `paymo docs` provides topic-based documentation (auth, time, projects, formats, examples)
- Command aliases exist (`projects` / `project` / `proj`)

**What's missing for AI agents:**
- **No machine-readable command listing.** An AI agent must parse human-readable `--help` text
  to discover commands. There's no `paymo --help --format json` or `paymo commands --json` to
  get structured command metadata.
- **No JSON schema for outputs.** An agent has no way to know the shape of JSON responses
  without making a call and inspecting the result.
- **No flag enumeration.** An agent can't programmatically discover what flags a command accepts.

### 3.2 Output Parsability — MIXED

**What works:**
- All list commands (`projects list`, `tasks list`, `time log`) support `--format json`
- JSON output uses proper tags and is parseable
- CSV output includes headers

**What's broken:**
- **Mutation commands don't support `--format json`.** Commands like `projects create`,
  `tasks create`, `time start`, `time stop`, `auth login` all produce human-formatted text
  with emoji:
  ```
  ✅ Project created successfully
     ID: 42
     Name: My Project
  ```
  An AI agent must regex-parse this to extract the created resource ID. This is the single
  biggest AI-usability problem.

- **Errors are always plain text.** Even with `--format json`, errors go to stderr as:
  ```
  Error: project not found
  ```
  Not as structured JSON. An AI agent can't distinguish error types programmatically.

- **JSON output lacks metadata envelope.** The spec called for
  `{"time_entries": [...], "total_count": 1}` but the implementation returns bare arrays
  `[{...}]`. An agent can't get a count without `len()`.

- **No `--quiet` flag.** Some commands print multiple lines of decorative output. There's no
  way to get just an ID or a minimal confirmation.

### 3.3 Error Handling for AI — WEAK

| Issue | Impact |
|-------|--------|
| All errors exit with code 1 | Agent can't distinguish auth failure from not-found from rate-limit |
| Errors are always text, never JSON | Agent must regex-parse error messages |
| Some errors lack corrective suggestions | "project not found" doesn't tell agent what projects exist |
| No error codes | No machine-readable error taxonomy |

**Recommended exit code scheme:**
- `0` — success
- `1` — general error
- `2` — usage/argument error
- `3` — authentication error
- `4` — not found
- `5` — rate limited
- `6` — network/API error

### 3.4 Consistency Issues Affecting AI Usage

- **ID handling is inconsistent.** `projects show` accepts both names and IDs, but `tasks show`
  and `tasks complete` accept only numeric IDs. An AI that learns "I can use names" from
  projects will fail on tasks.

- **`time log` vs `projects list`.** The "list" verb is used for projects and tasks, but
  time entries use `log`. An AI agent inferring patterns would try `time list`.

- **`--verbose` flag is defined but barely used.** Only one usage in `root.go`. An AI agent
  expecting debug output from `-v` gets nothing useful.

### 3.5 What Would Make This Truly AI-Friendly

**High priority:**
1. **JSON output for all commands.** Every mutation command should return the created/modified
   resource as JSON when `--format json` is used.
2. **Structured error output.** When `--format json`, errors should be JSON:
   ```json
   {"error": {"code": "NOT_FOUND", "message": "project not found", "status": 404}}
   ```
3. **Distinct exit codes** per error category (see table above).
4. **A `--quiet` flag** that suppresses decorative output and returns only essential data (an ID
   on create, nothing on delete).

**Medium priority:**
5. **Consistent ID-or-name resolution** across all commands.
6. **A `paymo schema` or `paymo commands --json` command** that outputs machine-readable command
   metadata for agent bootstrapping.
7. **Metadata envelope on JSON output** matching the spec: `{"data": [...], "count": N}`.

**Lower priority:**
8. **Shell completions** — not for AI, but for the humans who configure AI agents.
9. **`--raw` mode** — output a single value (e.g., just the entry ID after `time start`).
10. **Idempotency hints** — indicate which commands are safe to retry.

---

## 4. Architecture

### 4.1 Project Structure

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

### 4.2 Strengths

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

### 4.3 Weaknesses

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

## 5. Bugs and Code Issues

### 5.1 BUG: Nil pointer dereference in CSV output

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

### 5.2 CODE SMELL: Ignored errors in `fmt.Sscanf`

**File:** `internal/api/client.go:180-188`
**Severity:** Low — silent failure, rate limiting may not work correctly

```go
fmt.Sscanf(limit, "%d", &c.rateLimit)       // error ignored
fmt.Sscanf(remaining, "%d", &c.rateRemaining) // error ignored
fmt.Sscanf(decay, "%d", &seconds)            // error ignored
```

If the Paymo API returns a non-numeric value in these headers, parsing silently fails and rate
limiting may behave incorrectly. Use `strconv.Atoi` with explicit error handling instead.

### 5.3 INPUT SANITIZATION: Unsanitized user input in `where` clauses

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

## 6. Testing

### 6.1 Current State

| Package | Test Files | Approx. LOC | Coverage |
|---------|-----------|-------------|----------|
| `internal/api/` | 3 | ~710 | Partial (client, entries, projects) |
| `internal/config/` | 2 | ~295 | Good (config + timer) |
| `internal/output/` | 1 | ~296 | Good |
| `cmd/` | 0 | 0 | **None** |
| **Total** | **6** | **~1,300** | **~35% of files** |

### 6.2 What's Good

- **Mock HTTP servers.** API tests use `httptest.NewServer()` correctly, avoiding external
  dependencies.
- **Table-driven tests.** `config_test.go` and `timer_test.go` use idiomatic Go table-driven
  patterns.
- **Temp directories.** Tests use `t.TempDir()` for file isolation.
- **Both happy and error paths.** API tests verify successful responses and error conditions.

### 6.3 What's Missing

- **Command-layer tests (`cmd/`).** The entire CLI layer — argument parsing, flag validation,
  output selection, error messaging — is untested. This is the largest gap.
- **Task API tests.** `internal/api/tasks.go` has no dedicated test file (`tasks_test.go` does
  not exist).
- **Edge cases.** No tests for: empty results, unicode in names, very large datasets, concurrent
  timer operations, corrupted config files.
- **Integration / E2E tests.** No tests that exercise the full command → API → output pipeline.
- **Race condition tests.** No `-race` flag usage despite the rate limiter using a mutex.

### 6.4 Recommendations

1. Add `cmd/` tests using Cobra's test utilities (`cmd.SetArgs()`, `cmd.SetOut()`).
2. Add `tasks_test.go` mirroring the existing `projects_test.go` structure.
3. Run tests with `-race` flag to catch data races.
4. Add a CI pipeline that runs `go test -race -cover ./...` on every push.

---

## 7. Security

### 7.1 Positive Practices

| Area | Implementation | Assessment |
|------|---------------|------------|
| Credential storage | `~/.config/paymo-cli/credentials` with `0600` perms | Good |
| Config directory | Created with `0700` permissions | Good |
| Permission validation | `CheckCredentialsPermissions()` warns on loose perms | Good |
| Password handling | Not stored; read via `term.ReadPassword()` | Good |
| Transport | Default `https://` URL | Good |
| No shell execution | Pure HTTP client, no `exec.Command` usage | Good |
| No hardcoded secrets | Test files use dummy values only | Good |

### 7.2 Concerns

1. **Plaintext API key storage.** The API key in `credentials` is not encrypted. Acceptable for
   a CLI tool with proper file permissions, but worth documenting for users.

2. **No HTTPS enforcement.** The `base_url` config value can be overridden to `http://`. The
   client should reject non-HTTPS URLs or at minimum warn.

3. **Where-clause injection.** As noted in section 3.3, user input flows unsanitized into API
   query parameters.

4. **No dependency auditing.** No `govulncheck` or equivalent in the build process.

---

## 8. Code Quality & Developer Experience

### 8.1 Missing Tooling

- **No linting configuration.** No `.golangci-lint.yml` or equivalent. `go vet` is the only
  static analysis available.
- **No CI/CD pipeline.** No `.github/workflows/`, `Makefile`, or equivalent automation.
- **No `Makefile`.** Common Go project convention for `build`, `test`, `lint`, `install` targets.
- **No pre-commit hooks.** No automated checks before commits.

### 8.2 go.mod Issues

- **All dependencies marked `// indirect`.** The direct dependencies (`cobra`, `viper`, `term`)
  should not have the `// indirect` comment. This suggests the module was initialized incorrectly
  or `go mod tidy` was run in an unusual way.

- **Toolchain pinned to `go1.24.13`.** This specific patch version may not be available on all
  systems. Consider using just the minor version (`go 1.24`) or removing the `toolchain` directive.

### 8.3 Code Style

- Consistent formatting (likely `gofmt`-compliant).
- Good use of Go error wrapping with `%w`.
- Descriptive function and variable names.
- Comments on exported functions.
- One inconsistency: some API methods return `(*Type, error)` while others return `(Type, error)`.

---

## 9. Feature Completeness

See section 2 for the full spec-vs-implementation gap analysis. Summary of what's implemented:

| Feature | Status | Notes |
|---------|--------|-------|
| API key auth | Implemented | |
| Email/password auth | Implemented | |
| Time start/stop/status | Implemented | |
| Time log with filters | Implemented | |
| Project list/show/create/archive | Implemented | Missing: update |
| Task list/show/create/complete | Implemented | Missing: update |
| Table output | Implemented | |
| JSON output | Implemented | Missing metadata envelope per spec |
| CSV output | Implemented | Has nil-pointer bug (5.1) |
| Built-in docs | Implemented | |
| Man page generation | Implemented | |
| Rate limiting | Implemented | Has silent parse failures (5.2) |
| Offline timer | Implemented | Via local JSON file |
| Time edit/delete | **Not implemented** | Specified in SPECIFICATIONS.md |
| Reports module | **Not implemented** | Specified in SPECIFICATIONS.md |
| SQLite caching | **Not implemented** | Specified in SPECIFICATIONS.md |
| Offline sync | **Not implemented** | Specified in SPECIFICATIONS.md |
| Shell completions | **Not implemented** | Cobra supports this natively |
| Pagination | **Not implemented** | Will fail on large datasets |
| Session token auth | **Not implemented** | Specified in SPECIFICATIONS.md |

---

## 10. Prioritized Recommendations

### P0 — Fix Now

1. **Fix nil pointer dereference** in `internal/output/output.go:299` — this is a crash bug.
2. **Add nil guards** everywhere `e.Task` or `e.Project` are dereferenced in output formatters.

### P1 — Fix Soon (AI-Friendliness)

3. **JSON output for mutation commands.** `projects create`, `tasks create`, `time start`,
   `time stop` all print emoji-decorated text. When `--format json` is set, these must return
   the created/modified resource as JSON. This is the single biggest blocker for AI usage.
4. **Structured error output.** When `--format json`, errors should be JSON on stderr:
   `{"error": {"code": "NOT_FOUND", "message": "...", "status": 404}}`.
5. **Distinct exit codes.** 0=success, 2=usage, 3=auth, 4=not-found, 5=rate-limit, 6=API error.
6. **Consistent ID-or-name resolution.** `tasks show` and `tasks complete` only accept numeric
   IDs, while `projects show` accepts names. Make all commands accept both.

### P1 — Fix Soon (Quality)

7. **Add `cmd/` test coverage.** The entire 1,800-line command layer is untested.
8. **Add `tasks_test.go`.** The only API resource without tests.
9. **Sanitize user input** in `where` clause construction (`projects.go:86`, `tasks.go:86`).
10. **Fix `go.mod`** — remove incorrect `// indirect` comments on direct dependencies.
11. **Add CI pipeline** — at minimum: `go vet`, `go test -race ./...`.

### P2 — Improve (AI-Friendliness)

12. **Add a `--quiet` flag.** Suppress decorative output; return only essential data (e.g.,
    just the entry ID after `time start`).
13. **JSON metadata envelope.** Match the spec: `{"data": [...], "count": N}` instead of
    bare arrays.
14. **Machine-readable command discovery.** A `paymo commands --json` or similar endpoint that
    lists all commands, their flags, and expected output shapes.
15. **Implement missing spec commands.** `time edit`, `time delete`, `projects update`,
    `tasks update` — an AI agent trying to perform full CRUD will hit dead ends.

### P2 — Improve (Quality)

16. **Add `golangci-lint` configuration** with sensible defaults.
17. **Add a `Makefile`** with standard targets (`build`, `test`, `lint`, `install`).
18. **Handle `fmt.Sscanf` errors** in rate limit parsing, or switch to `strconv.Atoi`.
19. **Add HTTPS enforcement** — reject or warn on `http://` base URLs.
20. **Add pagination** to list endpoints.
21. **Add shell completions** (`cobra.GenBashCompletion`, etc.).

### P3 — Nice to Have

22. **Add structured logging** behind the `--verbose` flag using `log/slog`.
23. **Define API layer interfaces** for better testability of the command layer.
24. **Add `govulncheck`** to CI for dependency security scanning.
25. **Implement SQLite cache** per spec — enables offline mode and reduces API calls.
26. **Idempotency annotations** — indicate which commands are safe to retry (useful for agents
    that encounter transient errors).

---

## 11. Conclusion

paymo-cli has solid architectural foundations — clean separation of concerns, good Cobra/Viper
usage, and proper security practices. However, it falls short in two critical dimensions:

**Spec conformance:** Roughly half the specified features are unimplemented, including the
entire caching/offline layer, the reports module, and several CRUD operations (edit, delete,
update). The JSON output format deviates from the spec's metadata-envelope design.

**AI-friendliness:** Despite the README's claim, the tool is currently **human-first, not
AI-first**. The core issue is that mutation commands (`create`, `start`, `stop`) output
emoji-decorated text instead of structured JSON, errors are always plain text with a single
exit code, and command discovery requires parsing human-readable help. An AI agent using this
tool today would need regex parsing for half its operations.

The good news is that the architectural foundations make these problems fixable. The formatter
abstraction already exists — mutation commands just need to use it. Error types already carry
status codes internally — they just need to be surfaced. The P1 AI-friendliness items (JSON
mutation output, structured errors, distinct exit codes, consistent ID resolution) would
transform this from "usable by AI with workarounds" to "designed for AI."
