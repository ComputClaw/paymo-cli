# CLAUDE.md — paymo-cli

## Project overview

CLI client for the Paymo time tracking / project management API. Written in Go 1.24 using Cobra + Viper. Module: `github.com/ComputClaw/paymo-cli`.

## Repository structure

```
cmd/paymo/main.go       # Entrypoint — version injected via ldflags
cmd/                    # Cobra commands (one file per resource)
  root.go               # Root command, global flags, viper bindings
  helpers.go            # Shared resolvers (resolveProject, resolveTask, etc.)
  time.go               # time start/stop/status/log/show/edit/delete
  projects.go           # projects list/show/create/archive/tasks
  tasks.go              # tasks list/show/create/complete
  clients.go            # clients list
  auth.go, cache.go, sync.go, schema.go, docs.go, ...
  cmd_test.go           # All command tests (mock-based, single file)
internal/
  api/
    interface.go        # PaymoAPI interface — both Client and CachedClient implement it
    client.go           # HTTP client (API key or basic auth)
    models.go           # Request/response structs
    entries.go          # Time entry API methods
    projects.go         # Project API methods
    tasks.go            # Task API methods
    clients.go          # Client (customer) API methods
  cache/
    cache.go            # BoltDB-based cache store
    cached_client.go    # CachedClient wrapping PaymoAPI with caching
    keys.go             # Cache key generation
  config/
    config.go           # Credentials, config file handling
    timer.go            # Local timer state (start/stop tracking)
  output/
    output.go           # Formatter — table, JSON, CSV output
docs/index.md           # AI agent guide (published to GitHub Pages)
.goreleaser.yml         # GoReleaser config for cross-platform releases
.github/workflows/
  ci.yml                # Build + test on push/PR to main
  release.yml           # GoReleaser on tag push (v*)
```

## Build and test

```bash
go build ./...          # Compile all packages
go test ./...           # Run all tests
go test -race -cover ./...  # CI runs this
go run ./cmd/paymo      # Run locally (version shows "dev")
```

Go binary is at `C:\Program Files\Go\bin\go.exe` on this Windows machine. Use PowerShell to invoke it if bash can't find it:
```
powershell -Command "& 'C:\Program Files\Go\bin\go.exe' test ./... 2>&1"
```

## Key patterns

### Adding a new command
1. Add the `*cobra.Command` variable in the appropriate `cmd/*.go` file
2. Register it in `init()` with `parentCmd.AddCommand(...)`
3. Use `getAPIClient()` to get the API client (mockable via function variable)
4. Use `newFormatter()` for output (respects `--format`, `--quiet` flags)
5. Use `formatter.FormatXxx()` methods for output — they handle JSON/table/CSV/quiet
6. Use `cmd.Flags().Changed("flagname")` to detect which flags were explicitly set

### Adding tests
- All command tests live in `cmd/cmd_test.go`
- Use `mockPaymoAPI` struct — implements `api.PaymoAPI` interface
- Use `runCommand(mock, "resource", "action", ...)` helper
- Reset any new command flags in `runCommand()` via `resetCommandFlags()`
- Tests run in JSON format (`viper.Set("format", "json")`)

### Resolvers
- `resolveProject(client, arg)` — accepts ID or name, returns `*api.Project`
- `resolveProjectID(client, arg)` — accepts ID or name, returns `int`
- `resolveTask(client, arg, projectFlag)` — accepts ID or name (name requires project context)

### API interface
All API methods are defined in `internal/api/interface.go` (`PaymoAPI` interface). Both the raw `Client` and `CachedClient` implement it. When adding new API methods:
1. Add to the interface
2. Implement in the raw client
3. Add caching wrapper in `cached_client.go`
4. Add to mock in `cmd/cmd_test.go`

### Releases
- Tag with `vN.N.N` and push — GoReleaser handles the rest
- Version is injected via ldflags: `-X main.version={{.Version}}`
- Release notes header configured in `.goreleaser.yml`

## Conventions
- Commit messages: imperative, concise (e.g. "Add time show command")
- No `docs:` / `test:` / `chore:` prefixes (GoReleaser filters those from changelogs)
- Update `docs/index.md` quick reference when adding user-facing commands
- Error wrapping: `fmt.Errorf("doing thing: %w", err)`
- Args validation: use `cobra.ExactArgs(n)` where possible
