# Architecture

## Technology Stack
- **Language**: Go 1.24+
- **CLI Framework**: Cobra + Viper
- **HTTP Client**: net/http with custom wrapper
- **Local Storage**: BoltDB-based cache (spec originally called for SQLite)
- **Configuration**: YAML with hierarchical precedence
- **Distribution**: Single binary via GitHub releases

## Project Structure
```
paymo-cli/
├── cmd/paymo/main.go       # Entrypoint — version injected via ldflags
├── cmd/                     # Cobra commands (one file per resource)
│   ├── root.go             # Root command, global flags, viper bindings
│   ├── helpers.go          # Shared resolvers (resolveProject, resolveTask)
│   ├── time.go             # time start/stop/status/log/show/edit/delete
│   ├── projects.go         # projects list/show/create/archive/tasks
│   ├── tasks.go            # tasks list/show/create/complete
│   ├── clients.go          # clients list
│   ├── auth.go             # auth login/logout/status
│   ├── cache.go            # cache status/clear
│   ├── sync.go             # sync command
│   ├── schema.go           # Machine-readable command schema
│   ├── docs.go             # Built-in documentation viewer
│   ├── man.go              # Man page generation
│   ├── completion.go       # Shell completions
│   └── cmd_test.go         # All command tests (mock-based)
├── internal/
│   ├── api/
│   │   ├── interface.go    # PaymoAPI interface
│   │   ├── client.go       # HTTP client (API key or basic auth)
│   │   ├── models.go       # Request/response structs
│   │   ├── entries.go      # Time entry API methods
│   │   ├── projects.go     # Project API methods
│   │   ├── tasks.go        # Task API methods
│   │   ├── clients.go      # Client (customer) API methods
│   │   └── me.go           # Current user endpoint
│   ├── cache/
│   │   ├── cache.go        # BoltDB-based cache store
│   │   ├── cached_client.go # CachedClient wrapping PaymoAPI
│   │   └── keys.go         # Cache key generation
│   ├── config/
│   │   ├── config.go       # Credentials, config file handling
│   │   └── timer.go        # Local timer state (start/stop tracking)
│   └── output/
│       └── output.go       # Formatter — table, JSON, CSV output
├── docs/index.md           # AI agent guide (GitHub Pages)
├── .goreleaser.yml         # Cross-platform release config
└── .github/workflows/
    ├── ci.yml              # Build + test on push/PR
    └── release.yml         # GoReleaser on tag push (v*)
```

## API Client Design

```go
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    Auth       AuthMethod
    RateLimit  *RateLimiter
    Cache      *Cache
}

type AuthMethod interface {
    SetAuth(req *http.Request) error
    Validate() error
}
```

### Authentication Methods
1. **API Key (Recommended)** — stored in config file with `0600` perms
2. **Email/Password** — interactive prompt with hidden password input

### Error Handling Strategy
- **Rate Limiting**: Respect `X-Ratelimit-*` headers, implement backoff
- **Network Errors**: Graceful fallback to cached data when offline
- **Authentication Errors**: Clear error messages with re-auth prompts
- **API Errors**: Parse Paymo error responses, provide actionable messages

## Configuration

### Precedence (highest to lowest)
1. Command-line flags
2. Environment variables (`PAYMO_*`)
3. Config file (`~/.config/paymo-cli/config.yaml`)
4. Built-in defaults

### Cache Strategy
- **Cache TTL**: Varies by data type (projects longer, entries shorter)
- **Offline Mode**: Fall back to stale cache when API unavailable
- **Invalidation**: Smart cache invalidation on mutations (create/update/delete)
- **Bypass**: `--no-cache` flag forces fresh API calls

## Testing Strategy

### Unit Tests
- API client methods (mock HTTP servers via `httptest`)
- Configuration parsing
- Output formatting
- Cache operations

### Command Tests
- All command tests in single `cmd/cmd_test.go`
- Mock API via `mockPaymoAPI` implementing `PaymoAPI` interface
- `runCommand()` helper for executing commands with mock
- Reset flags between tests via `resetCommandFlags()`

### Integration Tests
- Full command → API → output pipeline
- Error scenarios
- Offline/online transitions
