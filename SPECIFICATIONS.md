# Paymo CLI Specifications

## 📋 Project Overview

### Vision
Create a comprehensive command-line interface for Paymo that enables developers and power users to efficiently manage time tracking, projects, and tasks without leaving their terminal.

### Market Opportunity
- **First-mover advantage**: No existing CLI tools in the Paymo ecosystem
- **Large addressable market**: Paymo has extensive user base with no CLI solution
- **Proven demand**: Competitors like Toggl have multiple community CLI implementations
- **Developer-friendly**: Paymo's API is comprehensive and well-designed for automation

## 🎯 Core Objectives

1. **Primary Use Case**: Seamless time tracking integration into developer workflows
2. **Secondary Use Cases**: Project management, task handling, reporting
3. **User Experience**: Simple, fast, intuitive commands with sensible defaults
4. **Integration**: Shell completion, status in prompt, Git hooks potential
5. **Reliability**: Offline capability with sync, robust error handling

## 🏗️ Technical Architecture

### Technology Stack
- **Language**: Go 1.24+
- **CLI Framework**: Cobra + Viper
- **HTTP Client**: net/http with custom wrapper
- **Local Storage**: SQLite for caching and offline support
- **Configuration**: YAML with hierarchical precedence
- **Distribution**: Single binary via Homebrew, GitHub releases

### Project Structure
```
paymo-cli/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go
│   ├── time.go            # Time tracking commands
│   ├── projects.go        # Project management
│   ├── tasks.go           # Task management
│   ├── auth.go            # Authentication
│   └── reports.go         # Reporting
├── internal/
│   ├── api/               # Paymo API client
│   │   ├── client.go
│   │   ├── auth.go
│   │   ├── time.go
│   │   ├── projects.go
│   │   └── models.go
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   └── auth.go
│   ├── cache/             # Local SQLite cache
│   │   ├── cache.go
│   │   ├── models.go
│   │   └── migrations.go
│   └── output/            # Output formatting
│       ├── table.go
│       ├── json.go
│       └── csv.go
├── pkg/
│   └── models/            # Shared data models
│       ├── project.go
│       ├── task.go
│       ├── time.go
│       └── user.go
├── scripts/               # Build and release scripts
├── docs/                  # Documentation
├── main.go
├── go.mod
└── README.md
```

## 🔌 API Integration Specifications

### Authentication Methods
1. **API Key (Recommended)**
   - Store in secure config file
   - Used as username with any password
   - Format: `curl -u API_KEY:PLACEHOLDER`

2. **Email/Password**
   - Interactive prompt with hidden password input
   - Store hashed credentials securely

3. **Session Tokens**
   - For extended interactive sessions
   - Automatic renewal before expiration

### API Client Design
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

### Error Handling Strategy
- **Rate Limiting**: Respect `X-Ratelimit-*` headers, implement backoff
- **Network Errors**: Graceful fallback to cached data when offline
- **Authentication Errors**: Clear error messages with re-auth prompts
- **API Errors**: Parse Paymo error responses, provide actionable messages

## 📱 Command Line Interface Specification

### Command Structure Philosophy
- **Verb-Noun Pattern**: `paymo [verb] [noun] [options]`
- **Intuitive Grouping**: Logical command hierarchies
- **Progressive Disclosure**: Simple commands by default, power features via flags
- **Consistent Output**: Standardized formatting across all commands

### Core Command Groups

#### 1. Time Tracking (`paymo time`)
```bash
# Primary commands
paymo time start [project] [task] [description]
paymo time stop
paymo time status
paymo time log [filters]

# Advanced commands
paymo time edit <id> [fields]
paymo time delete <id>
paymo time sync
```

**Implementation Priority**: MVP - these are the most critical commands

#### 2. Projects (`paymo projects`)
```bash
# List and view
paymo projects list [filters]
paymo projects show <project>

# Management
paymo projects create <name> [options]
paymo projects update <project> [fields]
paymo projects archive <project>
```

#### 3. Tasks (`paymo tasks`)
```bash
# List and view
paymo tasks list [filters]
paymo tasks show <task>

# Management
paymo tasks create <name> [options]
paymo tasks update <task> [fields]
paymo tasks complete <task>
```

#### 4. Authentication (`paymo auth`)
```bash
paymo auth login [options]
paymo auth logout
paymo auth status
paymo auth refresh
```

#### 5. Reports (`paymo reports`)
```bash
paymo reports time [date-range] [filters]
paymo reports projects [filters]
paymo reports export <format> [options]
```

### Flag Design Principles

#### Global Flags
- `--verbose, -v`: Debug output
- `--format, -f`: Output format (table|json|csv)
- `--config`: Custom config file
- `--no-cache`: Skip cache, force API calls

#### Command-Specific Flags
- **Time Start**: `--project, -p`, `--task, -t`, `--description, -d`
- **Lists**: `--limit, -l`, `--active, -a`, `--sort, -s`
- **Filters**: `--date`, `--client`, `--user`

## 💾 Configuration Management

### Configuration Hierarchy (highest to lowest precedence)
1. Command-line flags
2. Environment variables (`PAYMO_*`)
3. Config file (`~/.paymo.yaml`)
4. Built-in defaults

### Configuration Schema
```yaml
# API Configuration
api:
  endpoint: "https://app.paymoapp.com/api"
  timeout: 30s
  rate_limit: 100  # requests per minute

# Authentication
auth:
  method: "api_key"  # api_key, email_password, session
  api_key: "${PAYMO_API_KEY}"
  email: ""
  
# Default Values
defaults:
  format: "table"
  project_id: ""
  timezone: "UTC"
  
# Output Configuration  
output:
  date_format: "2006-01-02"
  time_format: "15:04"
  table_style: "default"
  
# Cache Configuration
cache:
  enabled: true
  ttl: "1h"
  max_size: "10MB"
  location: "~/.cache/paymo"
```

## 🗄️ Local Storage & Caching

### SQLite Schema Design
```sql
-- Projects cache
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    client_id INTEGER,
    active BOOLEAN,
    billable BOOLEAN,
    created_at DATETIME,
    updated_at DATETIME,
    cached_at DATETIME,
    UNIQUE(id)
);

-- Tasks cache
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    project_id INTEGER,
    complete BOOLEAN,
    billable BOOLEAN,
    due_date DATE,
    cached_at DATETIME,
    UNIQUE(id)
);

-- Time entries cache
CREATE TABLE time_entries (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    task_id INTEGER,
    user_id INTEGER,
    start_time DATETIME,
    end_time DATETIME,
    duration INTEGER,
    description TEXT,
    billable BOOLEAN,
    cached_at DATETIME,
    UNIQUE(id)
);

-- Current timer state
CREATE TABLE timer_state (
    id INTEGER PRIMARY KEY DEFAULT 1,
    active BOOLEAN DEFAULT FALSE,
    project_id INTEGER,
    task_id INTEGER,
    description TEXT,
    start_time DATETIME,
    CHECK (id = 1)
);
```

### Cache Strategy
- **Cache TTL**: 1 hour default, configurable
- **Offline Mode**: Fall back to cache when API unavailable
- **Sync Strategy**: Background sync on command execution
- **Invalidation**: Smart cache invalidation on mutations

## 🎨 Output Formatting

### Table Format (Default)
```
┌─────┬──────────────────┬─────────────┬──────────┬───────────┐
│ ID  │ Project          │ Task        │ Duration │ Date      │
├─────┼──────────────────┼─────────────┼──────────┼───────────┤
│ 123 │ paymo-cli        │ Development │ 2h 30m   │ 2026-02-07│
│ 124 │ Client Project   │ Bug Fixing  │ 1h 15m   │ 2026-02-07│
└─────┴──────────────────┴─────────────┴──────────┴───────────┘
```

### JSON Format
```json
{
  "time_entries": [
    {
      "id": 123,
      "project": {
        "id": 456,
        "name": "paymo-cli"
      },
      "task": {
        "id": 789,
        "name": "Development"
      },
      "duration": 9000,
      "date": "2026-02-07",
      "description": "Working on CLI features"
    }
  ],
  "total_duration": 9000,
  "total_count": 1
}
```

### CSV Format
```csv
id,project_id,project_name,task_id,task_name,duration,date,description
123,456,paymo-cli,789,Development,9000,2026-02-07,"Working on CLI features"
```

## 🔄 Development Phases

### Phase 1: MVP (Week 1-2)
- [x] CLI structure with Cobra
- [x] Basic command scaffolding
- [ ] Authentication implementation
- [ ] Basic API client
- [ ] Time tracking (start/stop/status)
- [ ] Simple output formatting

### Phase 2: Core Features (Week 3-4)
- [ ] Project and task management
- [ ] Time entry listing and editing
- [ ] Local SQLite caching
- [ ] Configuration management
- [ ] Error handling and retries

### Phase 3: Advanced Features (Week 5-6)
- [ ] Offline mode with sync
- [ ] Report generation
- [ ] Multiple output formats
- [ ] Shell completions
- [ ] Interactive prompts

### Phase 4: Polish & Distribution (Week 7-8)
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Homebrew formula
- [ ] GitHub releases automation

## 🧪 Testing Strategy

### Unit Tests
- API client methods
- Configuration parsing
- Output formatting
- Cache operations

### Integration Tests
- Full command execution
- API interaction (with mock server)
- Configuration file handling
- Error scenarios

### End-to-End Tests
- Authentication flows
- Time tracking workflows
- Project management operations
- Offline/online transitions

## 📦 Distribution Strategy

### Release Channels
1. **GitHub Releases**: Primary distribution with automated builds
2. **Homebrew**: macOS and Linux package management
3. **Direct Download**: Standalone binaries for all platforms
4. **Package Managers**: Future: Chocolatey (Windows), Snap, AUR

### Build Pipeline
```yaml
# .github/workflows/release.yml
- Build for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Create GitHub release with assets
- Update Homebrew formula
- Generate checksums
- Sign binaries (future)
```

## 🔒 Security Considerations

### Credential Storage
- **API Keys**: Stored in config file with restricted permissions (600)
- **Passwords**: Never stored in plaintext, use session tokens
- **Config Security**: Warn if config file has loose permissions

### Network Security
- **TLS**: Enforce HTTPS for all API communication
- **Certificate Validation**: Strict certificate checking
- **Timeout**: Reasonable timeouts to prevent hanging

## 📈 Success Metrics

### User Experience
- **Command Execution Time**: < 200ms for cached operations
- **Error Recovery**: Clear error messages with suggested actions
- **Learning Curve**: New users productive within 5 minutes

### Technical Performance
- **API Rate Limits**: Stay well under Paymo's limits
- **Cache Hit Rate**: > 80% for frequent operations
- **Offline Capability**: Full functionality for recently accessed data

### Adoption Metrics
- GitHub stars and downloads
- Community contributions
- User feedback and feature requests

## 🔮 Future Enhancements

### Shell Integration
- Status display in shell prompt (Starship, Zsh)
- Git hooks for automatic time tracking
- Terminal notifications for running timers

### Advanced Features
- Project templates
- Bulk operations
- Custom reporting
- Team collaboration features
- Integration with other tools (Slack, Discord)

### Platform Expansion
- Web UI for complex operations
- Mobile companion app
- Browser extensions

---

This specification serves as the technical blueprint for developing a best-in-class CLI tool for the Paymo ecosystem. The focus is on developer experience, reliability, and leveraging the comprehensive Paymo API to create workflows that are impossible with the web interface alone.