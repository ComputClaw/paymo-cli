---
title: paymo-cli
description: A command-line client for Paymo time tracking and project management
---

# paymo-cli

A powerful command-line client for [Paymo](https://www.paymoapp.com/) time tracking and project management. Built for developers, freelancers, and AI agents who prefer the terminal.

## Installation

Download the latest binary for your platform from the [releases page](https://github.com/mbundgaard/paymo-cli/releases), extract it, and add it to your PATH.

**Windows (PowerShell):**

```powershell
$v = (Invoke-RestMethod "https://api.github.com/repos/mbundgaard/paymo-cli/releases/latest").tag_name.TrimStart("v")
Invoke-WebRequest -Uri "https://github.com/mbundgaard/paymo-cli/releases/latest/download/paymo-cli_${v}_windows_amd64.zip" -OutFile "$env:TEMP\paymo.zip"
Expand-Archive "$env:TEMP\paymo.zip" -DestinationPath "$env:LOCALAPPDATA\paymo-cli" -Force
# Add to PATH (run once)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:LOCALAPPDATA\paymo-cli", "User")
```

**macOS / Linux:**

```bash
VERSION=$(curl -s https://api.github.com/repos/mbundgaard/paymo-cli/releases/latest | grep '"tag_name"' | cut -d'"' -f4 | tr -d v)
# macOS Apple Silicon: darwin_arm64 | macOS Intel: darwin_amd64 | Linux: linux_amd64
curl -sL "https://github.com/mbundgaard/paymo-cli/releases/latest/download/paymo-cli_${VERSION}_darwin_arm64.tar.gz" | tar xz -C /usr/local/bin paymo
```

## Getting started

On first run, paymo-cli will walk you through setup:

1. **Configure** — You'll be prompted for your Paymo API key (find it in Paymo Settings > API Keys)
2. **Sync** — The CLI automatically downloads users, clients, and projects to the local cache so subsequent commands are fast and work offline

```bash
# First run — prompts for API key and syncs your data
paymo auth login

# Or provide the key directly
paymo auth login --api-key YOUR_API_KEY
```

Once configured, you're ready to go:

```bash
# Log a time entry
paymo time log --project "My Project" --task "Development" --date today --duration 2h -d "Feature work"

# View today's entries
paymo time list --date today

# Start a live timer
paymo time start "My Project" "Development"
```

## Features

### Time entry management

Adding and managing time entries is the core of paymo-cli. Log time against projects and tasks, edit entries, and review your timesheet — all from the terminal.

```bash
paymo time log --project <name> --task <name> --duration <duration> [-d "description"]
paymo time list [--date today] [--project NAME]
```

Date filters: `today`, `yesterday`, `this-week`, `last-week`, or `YYYY-MM-DD`.

### Live timer

For real-time tracking, start and stop a timer that persists across sessions.

```bash
paymo time start <project> <task> [-d "description"]
paymo time status
paymo time stop
```

### Project management

List, create, inspect, and archive projects.

```bash
paymo projects list [--active]
paymo projects show <name-or-id>
paymo projects create <name> [--client ID]
paymo projects archive <name-or-id>
```

### Task management

Full CRUD for tasks with completion tracking.

```bash
paymo tasks list --project <name-or-id>
paymo tasks show <task-id>
paymo tasks create <name> --project <id> --tasklist <id>
paymo tasks complete <task-id>
```

### Multiple output formats

All list commands support `--format table|json|csv`.

```bash
# Pretty Unicode table (default)
paymo projects list

# JSON for scripting and automation
paymo projects list --format json

# CSV for spreadsheets
paymo time log --format csv > timesheet.csv
```

### Sync and caching

After login, core data (user, clients, projects) is automatically synced into the local cache. You can also sync manually:

```bash
paymo sync                    # Sync core data: me, clients, projects
paymo sync all                # Sync everything including tasks
paymo sync projects clients   # Sync specific resources
paymo sync tasks              # Sync only tasks
```

Transparent caching layer reduces API calls and speeds up repeated queries. Manage it with:

```bash
paymo cache status
paymo cache clear
```

Use `--no-cache` on any command to bypass the cache.

### Shell completions

Generate completions for your shell:

```bash
paymo completion bash
paymo completion zsh
paymo completion fish
paymo completion powershell
```

### Built-in documentation

Access reference docs without leaving the terminal:

```bash
paymo docs              # List all topics
paymo docs auth         # Authentication help
paymo docs time         # Time tracking help
paymo docs formats      # Output format details
paymo docs examples     # Usage examples
```

## Configuration

Config files live in `~/.config/paymo-cli/`:

| File | Purpose |
|------|---------|
| `config.json` | API and output settings |
| `credentials.json` | Authentication (mode 0600) |
| `timer.json` | Active timer state |

### Environment variables

```bash
export PAYMO_API_KEY=your_key   # API key
export PAYMO_FORMAT=json        # Default output format
```

## API coverage

| Feature | Status |
|---------|--------|
| Authentication | API key + basic auth |
| Time Entries | Full CRUD |
| Projects | List, create, show, archive |
| Tasks | List, create, show, complete |
| Task Lists | List |
| Clients | List |
| Users | Current user info |
| Sync | Pre-populate cache on demand |
| Rate Limiting | Automatic handling |
| Caching | Transparent JSON file cache with TTL |

## Global flags

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Output format: table, json, csv |
| `--verbose` | `-v` | Verbose output |
| `--quiet` | `-q` | Minimal output (IDs only) |
| `--no-cache` | | Bypass cache, force fresh API calls |
| `--config` | | Custom config file path |

## Links

- [Source code](https://github.com/mbundgaard/paymo-cli)
- [Releases](https://github.com/mbundgaard/paymo-cli/releases)
- [Issues](https://github.com/mbundgaard/paymo-cli/issues)
- [Paymo API docs](https://github.com/paymo-org/api)
