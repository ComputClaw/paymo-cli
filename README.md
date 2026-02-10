# paymo-cli

A powerful command-line client for [Paymo](https://www.paymoapp.com/) time tracking and project management. Built for developers, freelancers, and AI agents who prefer the terminal.

## ✨ Features

- **Time Tracking**: Start/stop timers, log entries, view history with date filtering
- **Project Management**: List, create, show, and archive projects
- **Task Management**: Full CRUD for tasks with completion tracking
- **Multiple Output Formats**: Table (pretty Unicode), JSON, CSV
- **Local Timer State**: Timer persists across sessions — never lose tracked time
- **Sync & Caching**: Auto-sync on login, JSON file cache with TTL for fast offline access
- **Rate Limit Aware**: Respects Paymo API limits with automatic backoff
- **Built-in Documentation**: `paymo docs` for quick reference without leaving the terminal
- **AI-Friendly**: Consistent output formats and comprehensive `--help` for agent use

## 📦 Installation

### From Source

```bash
git clone https://github.com/ComputClaw/paymo-cli.git
cd paymo-cli
go build -o paymo .
sudo mv paymo /usr/local/bin/  # Optional: install globally
```

### Homebrew (Coming Soon)

```bash
brew install computclaw/tap/paymo-cli
```

### Binary Releases (Coming Soon)

Download pre-built binaries from the [releases page](https://github.com/ComputClaw/paymo-cli/releases).

## 🔧 Quick Start

### 1. Authenticate

```bash
# Using API key (recommended)
paymo auth login --api-key YOUR_API_KEY

# Check authentication status
paymo auth status
```

Get your API key from Paymo → Settings → API Keys.

### 2. Start Tracking Time

```bash
# Start a timer on a project/task
paymo time start "My Project" "Development" --description "Working on features"

# Check what's running
paymo time status

# Stop and save
paymo time stop
```

### 3. Explore

```bash
# List your projects
paymo projects list

# View today's time entries
paymo time log --date today

# Built-in docs
paymo docs
```

## 📖 Commands

### Time Tracking

```bash
paymo time start <project> <task> [-d "description"]  # Start timer
paymo time stop                                        # Stop and save
paymo time status                                      # Current timer status
paymo time log [--date DATE] [--project NAME]          # View entries
```

Date filters: `today`, `yesterday`, `this-week`, `last-week`, or `YYYY-MM-DD`

### Projects

```bash
paymo projects list [--active]              # List projects
paymo projects show <name-or-id>            # Project details
paymo projects create <name> [--client ID]  # Create project
paymo projects archive <name-or-id>         # Archive project
```

### Tasks

```bash
paymo tasks list --project <name-or-id>     # List tasks in project
paymo tasks show <task-id>                  # Task details
paymo tasks create <name> --project <id> --tasklist <id>  # Create task
paymo tasks complete <task-id>              # Mark complete
```

### Sync & Cache

```bash
paymo sync                        # Sync core data: me, clients, projects
paymo sync all                    # Sync everything including tasks
paymo sync projects tasks         # Sync specific resources
paymo cache status                # Cache statistics
paymo cache clear                 # Clear all cached data
```

### Authentication

```bash
paymo auth login [--api-key KEY]  # Authenticate (auto-syncs core data)
paymo auth status                 # Check current auth
paymo auth logout                 # Clear credentials
```

### Documentation

```bash
paymo docs                    # List all topics
paymo docs auth               # Authentication help
paymo docs time               # Time tracking help
paymo docs formats            # Output format details
paymo docs examples           # Usage examples
```

### Help & Info

```bash
paymo --version               # Show version
paymo --help                  # Global help
paymo help <command>          # Command-specific help
paymo man                     # Generate man pages
paymo markdown                # Generate markdown docs
```

## 📊 Output Formats

All list commands support multiple formats via `--format`:

```bash
# Pretty table (default)
paymo projects list

# JSON for scripting/automation
paymo projects list --format json

# CSV for spreadsheets
paymo time log --format csv > timesheet.csv
```

## ⚙️ Configuration

Configuration is stored in `~/.config/paymo-cli/`:

- `config.json` — API settings
- `credentials.json` — Authentication (mode 0600)
- `timer.json` — Active timer state

### Environment Variables

```bash
export PAYMO_API_KEY=your_key        # API key
export PAYMO_FORMAT=json             # Default output format
```

## 🔌 API Coverage

Built on the [Paymo REST API](https://github.com/paymo-org/api):

| Feature | Status |
|---------|--------|
| Authentication | ✅ API key + basic auth |
| Time Entries | ✅ Full CRUD |
| Projects | ✅ List, create, show, archive |
| Tasks | ✅ List, create, show, complete |
| Task Lists | ✅ List |
| Clients | ✅ List |
| Users | ✅ Current user info |
| Sync | ✅ Pre-populate cache on demand |
| Rate Limiting | ✅ Automatic handling |
| Caching | ✅ Transparent JSON file cache with TTL |
| Filtering | ✅ Paymo `where` syntax |

## 🧪 Development

```bash
# Run tests
go test ./...

# Build
go build -o paymo .

# Run with verbose output
./paymo --verbose projects list
```

### Project Structure

```
├── cmd/           # Cobra commands (auth, time, projects, tasks, sync, docs)
├── internal/
│   ├── api/       # Paymo API client with rate limiting
│   ├── cache/     # JSON file cache with TTL and stale fallback
│   ├── config/    # Configuration and timer state management
│   └── output/    # Table, JSON, CSV formatters
├── main.go
└── go.mod
```

## 📝 License

MIT License — see [LICENSE](LICENSE) for details.

## 🔗 Links

- [Paymo API Documentation](https://github.com/paymo-org/api)
- [Issues](https://github.com/ComputClaw/paymo-cli/issues)
- [Releases](https://github.com/ComputClaw/paymo-cli/releases)

---

Built with ❤️ by [ComputClaw](https://github.com/ComputClaw)
