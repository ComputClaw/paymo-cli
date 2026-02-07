# paymo-cli

A command-line client for Paymo time tracking and project management.

## 🚀 Features

- **Time Tracking**: Start/stop timers, log time entries
- **Project Management**: List, create, and manage projects
- **Task Management**: Handle tasks and task lists
- **Reporting**: Generate time reports and export data
- **Multiple Auth Methods**: Email/password or API key authentication
- **Multiple Output Formats**: Table, JSON, CSV output
- **Offline Capability**: Local caching with sync (planned)

## 📦 Installation

### From Source (Development)

```bash
git clone https://github.com/ComputClaw/paymo-cli.git
cd paymo-cli
go build -o paymo
```

### Homebrew (Coming Soon)

```bash
brew install computclaw/tap/paymo-cli
```

### Direct Download (Coming Soon)

Download the latest binary from the [releases page](https://github.com/ComputClaw/paymo-cli/releases).

## 🔧 Setup

### 1. Authentication

**Using API Key (Recommended):**
```bash
paymo auth login --api-key YOUR_API_KEY
```

**Using Email/Password:**
```bash
paymo auth login
```

### 2. Configuration

Create a configuration file at `~/.paymo.yaml`:

```yaml
api:
  endpoint: "https://app.paymoapp.com/api"
  timeout: 30s

defaults:
  format: table
  project_id: ""
  
output:
  timezone: "Europe/Copenhagen"
  date_format: "2006-01-02"
```

## 🕒 Time Tracking

### Quick Start

```bash
# Start tracking time
paymo time start "My Project" "Development Task" "Working on CLI features"

# Check status
paymo time status

# Stop tracking
paymo time stop

# View recent entries
paymo time log
```

### Advanced Usage

```bash
# Start with specific project/task IDs
paymo time start -p 123 -t 456 -d "Bug fixes"

# View time entries for specific date
paymo time log --date 2026-02-07

# Filter by project
paymo time log --project "paymo-cli"

# Export as JSON
paymo time log --format json > timesheet.json
```

## 📁 Project Management

```bash
# List all projects
paymo projects list

# Show active projects only
paymo projects list --active

# Create a new project
paymo projects create "New Project" --client "Client Name" --billable

# View project details
paymo projects show "Project Name"
```

## 🔑 Authentication

```bash
# Check authentication status
paymo auth status

# Switch to API key authentication
paymo auth login --api-key YOUR_NEW_KEY

# Logout
paymo auth logout
```

## 📊 Output Formats

All list commands support multiple output formats:

```bash
# Table format (default)
paymo projects list

# JSON format
paymo projects list --format json

# CSV format
paymo projects list --format csv
```

## ⚙️ Configuration

### Global Flags

- `--verbose, -v`: Enable verbose output
- `--format, -f`: Output format (table, json, csv)
- `--config`: Custom config file path

### Environment Variables

All configuration can be overridden with environment variables:

```bash
export PAYMO_API_KEY=your_key
export PAYMO_FORMAT=json
export PAYMO_VERBOSE=true
```

## 🔌 API Integration

This CLI uses the [Paymo API v1](https://github.com/paymo-org/api) with the following features:

- ✅ Authentication (API key, email/password)
- ✅ Rate limiting compliance
- ✅ Error handling
- ✅ Project management
- ✅ Time entry management
- ✅ Task management
- ✅ User management
- ✅ Report generation

## 🏗️ Development Status

This project is in active development. Current status:

- ✅ CLI structure and commands
- ✅ Configuration management
- 🚧 Paymo API client implementation
- 🚧 Authentication flows
- 🚧 Time tracking functionality
- 🚧 Project management
- 🚧 Offline caching
- 🚧 Shell completions
- 🚧 Tests

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🔗 Links

- [Paymo API Documentation](https://github.com/paymo-org/api)
- [Project Issues](https://github.com/ComputClaw/paymo-cli/issues)
- [Releases](https://github.com/ComputClaw/paymo-cli/releases)

---

Built with ❤️ by [ComputClaw](https://github.com/ComputClaw)