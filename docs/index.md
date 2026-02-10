---
title: paymo-cli
description: A command-line client for Paymo time tracking and project management
---

# paymo-cli

A powerful command-line client for [Paymo](https://www.paymoapp.com/) time tracking and project management. Built for developers, freelancers, and AI agents who prefer the terminal.

## Installation

Download the latest binary for your platform, extract it, and add it to your PATH.

**Windows:**

```powershell
powershell -ExecutionPolicy Bypass -Command "irm https://mbundgaard.github.io/paymo-cli/install.ps1 | iex"
```

**macOS (Apple Silicon):**

```bash
curl -sL https://github.com/mbundgaard/paymo-cli/releases/latest/download/paymo-cli_darwin_arm64.tar.gz | tar xz -C /usr/local/bin paymo
```

**macOS (Intel):**

```bash
curl -sL https://github.com/mbundgaard/paymo-cli/releases/latest/download/paymo-cli_darwin_amd64.tar.gz | tar xz -C /usr/local/bin paymo
```

**Linux:**

```bash
curl -sL https://github.com/mbundgaard/paymo-cli/releases/latest/download/paymo-cli_linux_amd64.tar.gz | tar xz -C /usr/local/bin paymo
```

All binaries are available on the [releases page](https://github.com/mbundgaard/paymo-cli/releases).

## Getting started

After installation, authenticate with a Paymo API key. Ask the user for their key (they can find it in Paymo under Settings > Integrations > API keys).

```bash
paymo auth login --api-key <KEY>
```

This validates the key, stores credentials, and syncs core data (user, clients, projects) into the local cache.

Verify it worked:

```bash
paymo auth status
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
