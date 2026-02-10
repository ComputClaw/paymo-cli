# Roadmap

## Vision

Create a comprehensive command-line interface for Paymo that enables developers, power users,
and AI agents to manage time tracking, projects, and tasks without leaving their terminal.

### Core Objectives
1. **Primary Use Case**: Seamless time tracking integration into developer workflows
2. **Secondary Use Cases**: Project management, task handling, reporting
3. **User Experience**: Simple, fast, intuitive commands with sensible defaults
4. **AI-First**: Structured output, machine-readable discovery, consistent behavior
5. **Reliability**: Offline capability with cache fallback, robust error handling

## Completed

- [x] CLI structure with Cobra + Viper
- [x] API key and email/password authentication
- [x] Time tracking: start/stop/status/log/show/edit/delete
- [x] Project management: list/show/create/archive/tasks
- [x] Task management: list/show/create/complete
- [x] Client listing
- [x] Table, JSON, CSV output formats
- [x] JSON output for all commands including mutations
- [x] Structured JSON error output
- [x] `--quiet` flag for minimal output
- [x] BoltDB cache with stale fallback on network errors
- [x] `--no-cache` flag to bypass cache
- [x] Machine-readable command schema (`paymo schema`)
- [x] Shell completions
- [x] CI pipeline (vet + test with race detector)
- [x] GoReleaser release pipeline (cross-platform binaries)
- [x] AI agent guide (GitHub Pages)
- [x] Built-in documentation viewer (`paymo docs`)
- [x] Man page and markdown generation

## Prioritized Backlog

### P1 — High Value

1. **Distinct exit codes** — 0=success, 2=usage, 3=auth, 4=not-found, 5=rate-limit, 6=API error
2. **Sanitize user input** in `where` clause construction (`projects.go`, `tasks.go`)
3. **Fix `fmt.Sscanf` errors** in rate limit parsing — switch to `strconv.Atoi`
4. **`paymo projects update`** — update project fields
5. **`paymo tasks update`** — update task fields

### P2 — Important

6. **Pagination** on list endpoints — handle large datasets
7. **JSON metadata envelope** — `{"data": [...], "count": N}` instead of bare arrays
8. **HTTPS enforcement** — reject or warn on `http://` base URLs
9. **`paymo reports` command group** — time reports, project reports, export
10. **`govulncheck`** in CI for dependency security scanning
11. **`golangci-lint` configuration** with sensible defaults

### P3 — Nice to Have

12. **Structured logging** behind `--verbose` flag using `log/slog`
13. **`--sort` flag** on list commands
14. **`--user` filter** for multi-user setups
15. **Idempotency annotations** — indicate which commands are safe to retry
16. **Homebrew formula** for macOS/Linux distribution
17. **Makefile** with standard targets (build, test, lint, install)

## Future Enhancements

### Shell Integration
- Status display in shell prompt (Starship, Zsh)
- Git hooks for automatic time tracking
- Terminal notifications for running timers

### Advanced Features
- Project templates
- Bulk operations
- Custom reporting
- Team collaboration features

### Distribution
- Homebrew formula
- Chocolatey (Windows)
- Snap / AUR (Linux)

## Success Metrics

### User Experience
- Command execution time: < 200ms for cached operations
- Error recovery: clear error messages with suggested actions
- Learning curve: new users productive within 5 minutes

### Technical Performance
- API rate limits: stay well under Paymo's limits
- Cache hit rate: > 80% for frequent operations
- Offline capability: full functionality for recently accessed data
