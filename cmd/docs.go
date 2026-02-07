package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// docsCmd shows comprehensive documentation
var docsCmd = &cobra.Command{
	Use:   "docs [topic]",
	Short: "Show comprehensive documentation",
	Long: `Display detailed documentation and examples for paymo-cli.

Topics:
  auth       Authentication setup and methods
  time       Time tracking commands
  projects   Project management
  tasks      Task management
  formats    Output format options
  config     Configuration options
  examples   Common usage examples

Run 'paymo docs <topic>' for detailed information.
Run 'paymo docs examples' for a quick-start guide.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return showOverview()
		}
		
		switch args[0] {
		case "auth":
			return showAuthDocs()
		case "time":
			return showTimeDocs()
		case "projects":
			return showProjectsDocs()
		case "tasks":
			return showTasksDocs()
		case "formats":
			return showFormatsDocs()
		case "config":
			return showConfigDocs()
		case "examples":
			return showExamples()
		default:
			return fmt.Errorf("unknown topic: %s\nRun 'paymo docs' to see available topics", args[0])
		}
	},
}

func showOverview() error {
	fmt.Print(`
PAYMO-CLI DOCUMENTATION
=======================

paymo-cli is a command-line client for Paymo time tracking and project management.

QUICK START
-----------
1. Authenticate:     paymo auth login --api-key YOUR_API_KEY
2. List projects:    paymo projects list
3. Start tracking:   paymo time start "Project" "Task" "Description"
4. Check status:     paymo time status
5. Stop tracking:    paymo time stop
6. View time log:    paymo time log

COMMAND STRUCTURE
-----------------
paymo <command> <subcommand> [arguments] [flags]

MAIN COMMANDS
-------------
  auth       Manage authentication (login, logout, status)
  time       Time tracking (start, stop, status, log)
  projects   Project management (list, create, show, archive)
  tasks      Task management (list, show, create, complete)
  docs       Show this documentation

GLOBAL FLAGS
------------
  --config string   Custom config file path
  --format string   Output format: table, json, csv (default "table")
  --verbose         Enable verbose output
  --help            Show help for any command

CONFIGURATION
-------------
Config file: ~/.config/paymo-cli/config.json
API endpoint: https://app.paymoapp.com/api

For detailed help on a topic, run: paymo docs <topic>
For command-specific help, run: paymo <command> --help
`)
	return nil
}

func showAuthDocs() error {
	fmt.Print(`
AUTHENTICATION
==============

paymo-cli supports two authentication methods:

1. API KEY (RECOMMENDED)
------------------------
Get your API key from Paymo: My Account → API Keys

  paymo auth login --api-key YOUR_API_KEY

The API key is stored securely in ~/.config/paymo-cli/config.json

2. EMAIL/PASSWORD
-----------------
Interactive login (password not stored):

  paymo auth login
  Email: your@email.com
  Password: ********

Note: Basic auth requires re-login if session expires.

COMMANDS
--------
  paymo auth login              Interactive login
  paymo auth login -k KEY       Login with API key
  paymo auth status             Check authentication status
  paymo auth logout             Clear stored credentials

ENVIRONMENT VARIABLES
---------------------
  PAYMO_API_KEY                 API key (overrides config file)

EXAMPLE
-------
  $ paymo auth login --api-key abc123xyz
  🔐 Validating credentials... ✅
  🎉 Successfully authenticated as John Doe (john@example.com)
`)
	return nil
}

func showTimeDocs() error {
	fmt.Print(`
TIME TRACKING
=============

Track time on projects and tasks with start/stop commands.

COMMANDS
--------
  paymo time start    Start a new timer
  paymo time stop     Stop the running timer
  paymo time status   Show current timer status
  paymo time log      List time entries

START TIMER
-----------
  paymo time start <project> <task> [description]
  paymo time start -p <project> -t <task> -d <description>

  Arguments can be names (fuzzy matched) or IDs.

  Examples:
    paymo time start "Website Redesign" "Development" "Working on homepage"
    paymo time start -p 123 -t 456 -d "Bug fixes"
    paymo time start "My Project" "Code Review"

STOP TIMER
----------
  paymo time stop

  Stops the currently running timer and saves the entry.

CHECK STATUS
------------
  paymo time status

  Shows: project, task, description, start time, elapsed time.

VIEW TIME LOG
-------------
  paymo time log [flags]

  Flags:
    --date string      Filter by date (today, yesterday, this-week, YYYY-MM-DD)
    --project string   Filter by project name or ID
    --format string    Output format (table, json, csv)

  Examples:
    paymo time log                        # Today's entries
    paymo time log --date yesterday       # Yesterday
    paymo time log --date 2026-02-01      # Specific date
    paymo time log --project "Website"    # Filter by project
    paymo time log --format json          # JSON output
`)
	return nil
}

func showProjectsDocs() error {
	fmt.Print(`
PROJECT MANAGEMENT
==================

List, create, and manage projects.

COMMANDS
--------
  paymo projects list       List all projects
  paymo projects show       Show project details
  paymo projects create     Create a new project
  paymo projects tasks      List tasks in a project
  paymo projects archive    Archive a project

LIST PROJECTS
-------------
  paymo projects list [flags]

  Flags:
    --active        Show only active projects (default true)
    --all           Include inactive projects
    --client ID     Filter by client ID
    --format        Output format (table, json, csv)

  Examples:
    paymo projects list
    paymo projects list --all
    paymo projects list --format json

SHOW PROJECT
------------
  paymo projects show <project>

  Argument can be project name or ID.

  Examples:
    paymo projects show 123
    paymo projects show "Website Redesign"

CREATE PROJECT
--------------
  paymo projects create <name> [flags]

  Flags:
    --client ID        Assign to client
    --description      Project description
    --billable         Mark as billable (default true)

  Examples:
    paymo projects create "New Project"
    paymo projects create "Client Work" --client 456 --billable

LIST PROJECT TASKS
------------------
  paymo projects tasks <project> [flags]

  Flags:
    --all    Include completed tasks

  Examples:
    paymo projects tasks 123
    paymo projects tasks "Website" --all

ARCHIVE PROJECT
---------------
  paymo projects archive <project>

  Examples:
    paymo projects archive 123
    paymo projects archive "Old Project"
`)
	return nil
}

func showTasksDocs() error {
	fmt.Print(`
TASK MANAGEMENT
===============

List, create, and manage tasks.

COMMANDS
--------
  paymo tasks list       List tasks
  paymo tasks show       Show task details
  paymo tasks create     Create a new task
  paymo tasks complete   Mark task as complete

LIST TASKS
----------
  paymo tasks list [flags]

  Flags:
    --project string   Filter by project name or ID
    --all              Include completed tasks
    --format           Output format (table, json, csv)

  Examples:
    paymo tasks list
    paymo tasks list --project "Website"
    paymo tasks list --project 123 --all

SHOW TASK
---------
  paymo tasks show <task-id>

  Examples:
    paymo tasks show 456

CREATE TASK
-----------
  paymo tasks create <name> --project <project> [flags]

  Flags:
    --project string   Project name or ID (required)
    --description      Task description
    --billable         Mark as billable (default true)
    --due              Due date (YYYY-MM-DD)

  Examples:
    paymo tasks create "New Feature" --project 123
    paymo tasks create "Bug Fix" -p "Website" --due 2026-02-15

COMPLETE TASK
-------------
  paymo tasks complete <task-id>

  Examples:
    paymo tasks complete 456
`)
	return nil
}

func showFormatsDocs() error {
	fmt.Print(`
OUTPUT FORMATS
==============

All list commands support multiple output formats via --format flag.

FORMATS
-------
  table    Human-readable table with borders (default)
  json     JSON array for parsing/automation
  csv      CSV for spreadsheet import

EXAMPLES
--------
  paymo projects list --format table
  paymo projects list --format json
  paymo projects list --format csv > projects.csv

  paymo time log --format json | jq '.[] | .duration'

JSON OUTPUT
-----------
JSON format is ideal for:
- Piping to jq for filtering
- Integration with other tools
- Programmatic access from scripts
- AI agent parsing

Example JSON time entry:
{
  "id": 123,
  "task_id": 456,
  "project": {"id": 789, "name": "Website"},
  "task": {"id": 456, "name": "Development"},
  "duration": 3600,
  "description": "Working on feature"
}

CSV OUTPUT
----------
CSV format is ideal for:
- Spreadsheet import (Excel, Google Sheets)
- Data analysis
- Reporting

The first row contains column headers.
`)
	return nil
}

func showConfigDocs() error {
	fmt.Print(`
CONFIGURATION
=============

paymo-cli stores configuration in ~/.config/paymo-cli/

FILES
-----
  config.json     Credentials and user settings

CONFIG FILE STRUCTURE
---------------------
{
  "auth_type": "api_key",
  "api_key": "your-api-key",
  "user_id": 123,
  "user_name": "Your Name"
}

ENVIRONMENT VARIABLES
---------------------
All config can be overridden with environment variables:

  PAYMO_API_KEY       API key for authentication
  PAYMO_FORMAT        Default output format (table/json/csv)
  PAYMO_VERBOSE       Enable verbose output (true/false)

PRECEDENCE
----------
1. Command-line flags (highest)
2. Environment variables
3. Config file
4. Built-in defaults (lowest)

SECURITY
--------
- Config file is created with 0600 permissions (owner read/write only)
- API keys are stored in plain text - protect your config directory
- Use environment variables in CI/CD pipelines

CUSTOM CONFIG FILE
------------------
Use --config flag to specify a custom config file:

  paymo --config /path/to/config.json projects list
`)
	return nil
}

func showExamples() error {
	fmt.Print(`
COMMON EXAMPLES
===============

SETUP
-----
# Login with API key
paymo auth login --api-key YOUR_API_KEY

# Verify authentication
paymo auth status

DAILY TIME TRACKING
-------------------
# Start your day - begin tracking
paymo time start "Client Project" "Development" "Morning standup and coding"

# Check what you're tracking
paymo time status

# Stop for lunch
paymo time stop

# Resume after lunch
paymo time start "Client Project" "Development" "Afternoon coding session"

# End of day - stop and review
paymo time stop
paymo time log

WEEKLY REVIEW
-------------
# View this week's time
paymo time log --date this-week

# Export to CSV for reporting
paymo time log --date this-week --format csv > weekly-report.csv

# View by project
paymo time log --project "Client Project"

PROJECT MANAGEMENT
------------------
# List all active projects
paymo projects list

# See project details and tasks
paymo projects show "Client Project"
paymo projects tasks "Client Project"

# Create a new project
paymo projects create "New Client" --billable

TASK MANAGEMENT
---------------
# List tasks for a project
paymo tasks list --project "Client Project"

# Create a new task
paymo tasks create "Implement feature X" --project "Client Project"

# Mark task complete
paymo tasks complete 123

AUTOMATION / SCRIPTING
----------------------
# Get project list as JSON
paymo projects list --format json

# Parse with jq
paymo time log --format json | jq '.[].duration' | awk '{sum+=$1} END {print sum/3600 " hours"}'

# Use in shell scripts
if paymo time status 2>/dev/null | grep -q "Timer Running"; then
  echo "Timer is active"
fi
`)
	return nil
}

func init() {
	rootCmd.AddCommand(docsCmd)
}