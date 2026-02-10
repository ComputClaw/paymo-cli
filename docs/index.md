---
title: paymo-cli
description: AI agent guide for installing and using paymo-cli
---

# paymo-cli — AI Agent Guide

This document is for AI agents. You are expected to install, authenticate, and operate paymo-cli on behalf of the user. Do not ask the user to run commands themselves — you run them.

## Step 1: Install

Run the install command for the user's platform.

**Windows:**

```
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

## Step 2: Authenticate

Ask the user for their Paymo API key. They can find it in Paymo under **Settings > Integrations > API keys**.

Once you have the key, run:

```bash
paymo auth login --api-key <KEY>
```

Do **not** use `paymo auth login` without `--api-key` — interactive mode requires a terminal and will fail.

## Step 3: Explore available commands

Run `paymo --help` to see all available commands. Use `paymo <command> --help` for details on any command.

## Quick reference

```bash
# Projects
paymo projects list
paymo projects show <name-or-id>

# Tasks
paymo tasks list --project <name-or-id>
paymo tasks show <task-id>

# Time tracking
paymo time start <project> <task> [-d "description"]
paymo time status
paymo time stop
paymo time log [--date today] [--project NAME]

# Sync & cache
paymo sync                          # Sync core data
paymo sync all                      # Sync everything
paymo cache status
paymo cache clear

# Output formats (all list commands)
paymo projects list --format json   # json, table, or csv
```

## Global flags

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Output format: table, json, csv |
| `--quiet` | `-q` | Minimal output (IDs only) |
| `--no-cache` | | Bypass cache, force fresh API calls |

## Links

- [Source code & full documentation](https://github.com/mbundgaard/paymo-cli)
- [Releases](https://github.com/mbundgaard/paymo-cli/releases)
