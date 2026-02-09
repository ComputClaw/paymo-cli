# paymo-cli Remediation Plan

**Based on:** REVIEW.md (2026-02-09)
**Organized into:** 10 phases, ordered by dependency and priority

---

## Dependency Graph

```
Phase 0 (nil fix)
  │
Phase 1A (error types + exit codes)
  │
Phase 1B (JSON mutations + structured errors)   Phase 1C (ID/name resolution)
  │                                                │
Phase 1D (tests, sanitization, go.mod, CI)  ◄──────┘
  │
Phase 2A (--quiet + envelope)     Phase 2B (schema command)
  │                                │
Phase 2C (missing CRUD commands)  Phase 2D (lint, Makefile, fixes)
  │
Phase 3A-E (logging, interfaces, cache, etc.)
```

---

## Phase 0: Fix Nil Pointer Crash

**Goal:** Eliminate the confirmed crash bug in CSV output.
**Complexity:** S | **Priority:** P0

### `internal/output/output.go`

In `formatEntriesCSV` (line 281), line 299 dereferences `e.Task.ProjectID` without a nil check
despite `e.Task` being checked for nil just above.

**Fix:** Extract `projectID` safely before building the CSV row:

```go
projectID := 0
if e.Task != nil {
    projectID = e.Task.ProjectID
}
```

Then use `fmt.Sprintf("%d", projectID)` instead of `fmt.Sprintf("%d", e.Task.ProjectID)`.

### `internal/output/output_test.go`

Add `TestFormatEntriesCSV_NilTask` — exercise the CSV formatter with `TimeEntry` where both
`Task` and `Project` are nil. Verify no panic and the entry ID appears in output.

---

## Phase 1A: Structured Error Types and Distinct Exit Codes

**Goal:** Machine-readable error classification. AI consumers can distinguish error categories
via exit codes and (later, in 1B) structured JSON.
**Complexity:** M | **Priority:** P1
**Depends on:** Phase 0

### `internal/api/client.go`

1. **Extend `APIError`** (line 86) with a `Code` field:

```go
type APIError struct {
    StatusCode int
    Code       string // "AUTH_FAILED", "NOT_FOUND", "RATE_LIMITED", "USAGE_ERROR", "API_ERROR"
    Message    string
    Details    map[string]interface{}
}
```

2. **Add `classifyHTTPStatus` helper:**

```go
func classifyHTTPStatus(statusCode int) string {
    switch {
    case statusCode == 401 || statusCode == 403:
        return "AUTH_FAILED"
    case statusCode == 404:
        return "NOT_FOUND"
    case statusCode == 429:
        return "RATE_LIMITED"
    case statusCode == 400:
        return "USAGE_ERROR"
    default:
        return "API_ERROR"
    }
}
```

3. **Set `Code` in `Request`** (line 149-161) when creating `APIError`:

```go
apiErr.Code = classifyHTTPStatus(resp.StatusCode)
```

4. **Add `ExitCode()` method on `APIError`:**

```go
func (e *APIError) ExitCode() int {
    switch e.Code {
    case "USAGE_ERROR":  return 2
    case "AUTH_FAILED":  return 3
    case "NOT_FOUND":    return 4
    case "RATE_LIMITED": return 5
    default:             return 6
    }
}
```

### `cmd/root.go`

In `init()`, add after global flag registration:

```go
rootCmd.SilenceErrors = true
rootCmd.SilenceUsage = true
```

This gives `main.go` full control over error formatting (needed for Phase 1B).

### `main.go`

Replace the error handler with exit-code-aware dispatch:

```go
func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        var ec interface{ ExitCode() int }
        if errors.As(err, &ec) {
            os.Exit(ec.ExitCode())
        }
        os.Exit(1)
    }
}
```

### `internal/api/client_test.go`

Add tests for `classifyHTTPStatus` (401→AUTH_FAILED, 404→NOT_FOUND, etc.) and `ExitCode()`.

---

## Phase 1B: JSON Output for Mutations and Structured Error Output

**Goal:** When `--format json`, mutation commands return the resource as JSON and errors are
structured JSON on stderr. This is the single biggest AI-friendliness improvement.
**Complexity:** L | **Priority:** P1
**Depends on:** Phase 1A

### `internal/output/output.go`

1. **Add single-resource formatting methods:**

```go
func (f *Formatter) FormatProject(project *api.Project) error
func (f *Formatter) FormatTask(task *api.Task) error
func (f *Formatter) FormatTimeEntry(entry *api.TimeEntry) error
```

Each method: if JSON, serialize the resource directly; if table, produce a detail view
(extract the current `fmt.Printf` emoji blocks from `cmd/` into private `formatProjectDetail`,
`formatTaskDetail`, `formatEntryDetail` methods).

2. **Add `FormatSuccess` for commands with no resource to return** (archive, logout, complete):

```go
type SuccessResult struct {
    Status  string `json:"status"`
    Message string `json:"message"`
    ID      int    `json:"id,omitempty"`
}

func (f *Formatter) FormatSuccess(msg string, id int) error
```

3. **Add `FormatError` for structured error output:**

```go
type ErrorResult struct {
    Error ErrorDetail `json:"error"`
}
type ErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"status,omitempty"`
}

func (f *Formatter) FormatError(err error)
```

When format is JSON, writes `{"error": {...}}` to stderr. Otherwise writes `"Error: ..."`.

### `cmd/projects.go`

- **`createProjectCmd`** (line 99-104): Replace `fmt.Printf("✅ ...")` with
  `formatter.FormatProject(project)`.
- **`showProjectCmd`** (line 144-169): Replace dual JSON/table branches with
  `formatter.FormatProject(project)`.
- **`archiveProjectCmd`** (line 258): Replace with `formatter.FormatSuccess(...)`.

### `cmd/tasks.go`

- **`createTaskCmd`** (line 167-170): Replace with `formatter.FormatTask(task)`.
- **`showTaskCmd`** (line 90-112): Replace with `formatter.FormatTask(task)`.
- **`completeTaskCmd`** (line 201): Replace with `formatter.FormatSuccess(...)`.

### `cmd/time.go`

- **`startCmd`** (line 143-149): Replace with `formatter.FormatTimeEntry(entry)`.
- **`stopCmd`** (line 189-194): Replace with `formatter.FormatTimeEntry(entry)`.
- **`statusCmd`** (line 209-223): Serialize `TimerState` as JSON when format is JSON.

### `cmd/auth.go`

- **`loginCmd`** (line 81-93): When JSON, return `{"status":"authenticated","user_id":N,...}`.
- **`logoutCmd`**: When JSON, return `{"status":"ok","message":"..."}`.
- **`statusAuthCmd`**: When JSON, serialize credentials/user info.

### `main.go`

Update error handler to use `FormatError`:

```go
format := cmd.GetOutputFormat() // new helper reading viper
formatter := output.NewFormatter(format)
formatter.FormatError(err)
```

### `internal/output/output_test.go`

Add tests for all new methods in both JSON and table modes.

---

## Phase 1C: Consistent ID-or-Name Resolution

**Goal:** `tasks show` and `tasks complete` accept names (not just numeric IDs), matching
`projects show` behavior.
**Complexity:** S | **Priority:** P1
**Depends on:** None (can parallel with 1A/1B)

### New file: `cmd/helpers.go`

Extract the ID-or-name resolution pattern that's duplicated across commands:

```go
func resolveProjectID(client *api.Client, arg string) (int, error) {
    if id, err := strconv.Atoi(arg); err == nil {
        return id, nil
    }
    project, err := client.GetProjectByName(arg)
    if err != nil {
        return 0, fmt.Errorf("project not found: %w", err)
    }
    return project.ID, nil
}

func resolveTask(client *api.Client, arg string, projectFlag string) (*api.Task, error) {
    if id, err := strconv.Atoi(arg); err == nil {
        return client.GetTask(id)
    }
    if projectFlag == "" {
        return nil, fmt.Errorf("task name lookup requires --project flag (or use numeric ID)")
    }
    projectID, err := resolveProjectID(client, projectFlag)
    if err != nil {
        return nil, err
    }
    return client.GetTaskByName(projectID, arg)
}
```

### `cmd/tasks.go`

- **`showTaskCmd`** (line 80-83): Replace `strconv.Atoi` with `resolveTask(client, args[0], projectFlag)`. Add `--project/-p` flag.
- **`completeTaskCmd`** (line 186-189): Same change. Add `--project/-p` flag.
- Update `Use` from `"show <task-id>"` to `"show <task>"`.

### `cmd/projects.go`, `cmd/time.go`

Replace duplicated ID-or-name resolution code with calls to `resolveProjectID` from `helpers.go`.

---

## Phase 1D: Tests, Input Sanitization, go.mod, CI

**Goal:** Fill the major quality gaps.
**Complexity:** L | **Priority:** P1
**Depends on:** Phases 1A-1C (so tests cover new code)

### New file: `internal/api/tasks_test.go`

Mirror `projects_test.go` structure. Tests for:
- `GetTasks`, `GetTasks` with project filter
- `GetTask` by ID
- `GetTaskByName`, `GetTaskByName` not found
- `CreateTask` — verify POST body and response
- `CompleteTask` — verify PUT with `{"complete": true}`
- `GetTaskLists`

### New file: `internal/api/sanitize.go`

```go
func sanitizeWhereInput(s string) string {
    s = strings.ReplaceAll(s, `"`, `\"`)
    s = strings.ReplaceAll(s, `%`, `\%`)
    return s
}
```

### `internal/api/projects.go` (line 86)

```go
// Before:
params.Set("where", fmt.Sprintf("name like \"%%%s%%\"", name))
// After:
params.Set("where", fmt.Sprintf("name like \"%%%s%%\"", sanitizeWhereInput(name)))
```

### `internal/api/tasks.go` (line 86)

Same sanitization fix.

### `cmd/auth.go`

Change `getAPIClient` from a plain function to a `var` for test injection:

```go
var getAPIClient = func() (*api.Client, error) { ... }
```

### New files: `cmd/projects_test.go`, `cmd/tasks_test.go`, `cmd/time_test.go`, `cmd/auth_test.go`

Use Cobra test utilities with mock HTTP servers:

```go
func executeCommand(args ...string) (string, error) {
    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetErr(buf)
    rootCmd.SetArgs(args)
    err := rootCmd.Execute()
    return buf.String(), err
}
```

Override `getAPIClient` to return a client pointing at `httptest.NewServer`.

### `go.mod`

Remove `// indirect` from direct dependencies:
- `github.com/spf13/cobra v1.10.2`
- `github.com/spf13/viper v1.21.0`
- `golang.org/x/term v0.39.0`
- `gopkg.in/yaml.v3 v3.0.1`

Remove `toolchain go1.24.13` line (or change to `toolchain local`).

### New file: `.github/workflows/ci.yml`

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go vet ./...
      - run: go test -race -cover ./...
```

---

## Phase 2A: `--quiet` Flag and JSON Metadata Envelope

**Goal:** Minimal output mode for scripts/agents, and spec-conformant JSON structure.
**Complexity:** M | **Priority:** P2
**Depends on:** Phase 1B

### `cmd/root.go`

Add in `init()`:

```go
rootCmd.PersistentFlags().BoolP("quiet", "q", false, "minimal output (IDs only)")
viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
```

### `internal/output/output.go`

1. Add `Quiet bool` to `Formatter`.

2. In quiet mode, mutation methods output only the resource ID:
```go
if f.Quiet {
    fmt.Fprintf(f.Writer, "%d\n", resource.ID)
    return nil
}
```

3. Wrap list JSON output in metadata envelope:
```go
type ListEnvelope struct {
    Data  interface{} `json:"data"`
    Count int         `json:"count"`
}
```

`FormatProjects`, `FormatTasks`, `FormatTimeEntries` wrap their JSON output:
```go
case "json":
    return f.formatJSON(ListEnvelope{Data: projects, Count: len(projects)})
```

**Note:** This is a breaking change. JSON output changes from `[...]` to `{"data":[...],"count":N}`.

### All `cmd/*.go` files

Update formatter creation to include quiet:

```go
formatter := output.NewFormatter(viper.GetString("format"))
formatter.Quiet = viper.GetBool("quiet")
```

---

## Phase 2B: Machine-Readable Command Discovery

**Goal:** `paymo schema` outputs all commands, flags, and aliases as structured JSON.
**Complexity:** M | **Priority:** P2
**Depends on:** None

### New file: `cmd/schema.go`

```go
type CommandSchema struct {
    Name        string          `json:"name"`
    FullCommand string          `json:"full_command"`
    Description string          `json:"description"`
    Aliases     []string        `json:"aliases,omitempty"`
    Flags       []FlagSchema    `json:"flags,omitempty"`
    Subcommands []CommandSchema `json:"subcommands,omitempty"`
    Idempotent  bool            `json:"idempotent"`
}

type FlagSchema struct {
    Name      string `json:"name"`
    Shorthand string `json:"shorthand,omitempty"`
    Type      string `json:"type"`
    Default   string `json:"default,omitempty"`
    Usage     string `json:"usage"`
}
```

`buildSchema(cmd *cobra.Command)` walks the command tree recursively, extracting metadata
from each `cobra.Command` and its `Flags()`.

---

## Phase 2C: Missing CRUD Commands

**Goal:** Implement `time edit`, `time delete`, `projects update`, `tasks update` per spec.
**Complexity:** L | **Priority:** P2
**Depends on:** Phase 1B (for formatter integration)

### `internal/api/models.go`

Add `UpdateProjectRequest` and `UpdateTaskRequest` with pointer fields for partial updates:

```go
type UpdateProjectRequest struct {
    Name        *string  `json:"name,omitempty"`
    Description *string  `json:"description,omitempty"`
    Billable    *bool    `json:"billable,omitempty"`
    BudgetHours *float64 `json:"budget_hours,omitempty"`
}

type UpdateTaskRequest struct {
    Name        *string `json:"name,omitempty"`
    Description *string `json:"description,omitempty"`
    DueDate     *string `json:"due_date,omitempty"`
    Priority    *int    `json:"priority,omitempty"`
}
```

### `internal/api/projects.go`

Add `UpdateProject(id int, req *UpdateProjectRequest) (*Project, error)` — PUT to
`/projects/{id}`.

### `internal/api/tasks.go`

Add `UpdateTask(id int, req *UpdateTaskRequest) (*Task, error)` — PUT to `/tasks/{id}`.

### `cmd/projects.go`

Add `updateProjectCmd` with flags `--name`, `--description`, `--billable`, `--budget-hours`.
Register under `projectsCmd`.

### `cmd/tasks.go`

Add `updateTaskCmd` with flags `--name`, `--description`, `--due`, `--priority`.
Register under `tasksCmd`.

### `cmd/time.go`

Add `editCmd` — calls `client.UpdateEntry` (already exists in `entries.go:107`).
Flags: `--description`, `--duration`, `--start`, `--end`.

Add `deleteCmd` — calls `client.DeleteEntry` (already exists in `entries.go:121`).
Output via `formatter.FormatSuccess(...)`.

---

## Phase 2D: Lint, Makefile, Code Fixes, Shell Completions

**Goal:** Address remaining P2 quality items.
**Complexity:** M | **Priority:** P2
**Depends on:** None (independent)

### New file: `.golangci.yml`

Enable: `errcheck`, `govet`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, `misspell`,
`gosec`, `gocritic`.

### New file: `Makefile`

Targets: `build`, `test`, `lint`, `install`, `clean`.

### `internal/api/client.go`

**Fix `fmt.Sscanf`** (lines 180-188) — replace with `strconv.Atoi`:

```go
if v, err := strconv.Atoi(limit); err == nil {
    c.rateLimit = v
}
```

**HTTPS enforcement** — in `NewClientWithBaseURL`, warn if URL doesn't start with `https://`
(allow `http://localhost` and `http://127.0.0.1` for tests).

### `internal/api/projects.go`, `tasks.go`, `entries.go`

Add `Page` and `PerPage` fields to `*ListOptions` structs. Set `page` and `per_page` query
params when non-zero.

### New file: `cmd/completion.go`

Cobra built-in shell completions for bash, zsh, fish, powershell:

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate shell completion scripts",
    ...
}
```

### New file: `.github/workflows/ci.yml` (if not added in 1D)

Add `golangci-lint-action` step.

---

## Phase 3A: Structured Logging

**Complexity:** S | **Priority:** P3

Initialize `log/slog` in `initConfig()` based on `--verbose`. Add `slog.Debug` calls in
`client.go:Request()` for request/response logging.

## Phase 3B: API Layer Interface

**Complexity:** S | **Priority:** P3

New file `internal/api/interface.go` defining `PaymoAPI` interface. All methods already
implemented by `Client`. Enables clean dependency injection in `cmd/` tests.

## Phase 3C: `govulncheck` in CI

**Complexity:** S | **Priority:** P3

Add `govulncheck ./...` step to CI pipeline.

## Phase 3D: SQLite Cache

**Complexity:** L | **Priority:** P3

New package `internal/cache/` with SQLite schema from SPECIFICATIONS.md. Cache layer wraps
API client. `--no-cache` flag. `paymo time sync` command. Cache TTL and invalidation on
mutations. This is a multi-week effort.

## Phase 3E: Idempotency Annotations

**Complexity:** S | **Priority:** P3

Extend `CommandSchema` (from Phase 2B) with `Idempotent` and `SafeToRetry` fields.
Mark GET commands as idempotent, mutations as non-idempotent.

---

## Summary

| Phase | Description | Files Changed | Files Created | Size |
|-------|-------------|---------------|---------------|------|
| **0** | Nil pointer crash fix | 1 | 0 | S |
| **1A** | Error types + exit codes | 3 | 0 | M |
| **1B** | JSON mutations + structured errors | 7 | 0 | L |
| **1C** | Consistent ID/name resolution | 3 | 1 | S |
| **1D** | Tests, sanitization, go.mod, CI | 4 | 7 | L |
| **2A** | --quiet + JSON envelope | 8+ | 0 | M |
| **2B** | Machine-readable schema | 0 | 1 | M |
| **2C** | Missing CRUD commands | 5 | 0 | L |
| **2D** | Lint, Makefile, fixes, completions | 4 | 4 | M |
| **3A-E** | Logging, interfaces, cache | varies | varies | S-L |
