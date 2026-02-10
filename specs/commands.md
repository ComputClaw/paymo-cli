# CLI Command Specification

## Command Structure Philosophy
- **Verb-Noun Pattern**: `paymo [verb] [noun] [options]`
- **Intuitive Grouping**: Logical command hierarchies
- **Progressive Disclosure**: Simple commands by default, power features via flags
- **Consistent Output**: Standardized formatting across all commands

## Global Flags
- `--verbose, -v`: Debug output
- `--format, -f`: Output format (table|json|csv)
- `--config`: Custom config file
- `--no-cache`: Skip cache, force API calls
- `--quiet, -q`: Minimal output (IDs only for create/mutate commands)

## Command Groups

### 1. Time Tracking (`paymo time`)
```bash
# Primary commands
paymo time start [project] [task] [description]
paymo time stop
paymo time status
paymo time log [filters]

# Entry management
paymo time show <id>
paymo time edit <id> [--description "..."] [--duration 2h] [--task 456]
paymo time delete <id>
paymo time sync
```

**Command-Specific Flags:**
- **Time Start**: `--project, -p`, `--task, -t`, `--description, -d`
- **Time Log**: `--date`, `--project`
- **Time Edit**: `--description, -d`, `--duration`, `--task, -t`

### 2. Projects (`paymo projects`)
```bash
# List and view
paymo projects list [filters]
paymo projects show <project>

# Management
paymo projects create <name> [options]
paymo projects update <project> [fields]
paymo projects archive <project>
paymo projects tasks <project>
```

### 3. Tasks (`paymo tasks`)
```bash
# List and view
paymo tasks list [filters]
paymo tasks show <task>

# Management
paymo tasks create <name> [options]
paymo tasks update <task> [fields]
paymo tasks complete <task>
```

### 4. Clients (`paymo clients`)
```bash
paymo clients list
```

### 5. Authentication (`paymo auth`)
```bash
paymo auth login [options]
paymo auth logout
paymo auth status
paymo auth refresh
```

### 6. Reports (`paymo reports`)
```bash
paymo reports time [date-range] [filters]
paymo reports projects [filters]
paymo reports export <format> [options]
```

### 7. Utility Commands
```bash
paymo sync                  # Sync core data
paymo sync all              # Sync everything
paymo cache status
paymo cache clear
paymo schema                # Machine-readable command schema (JSON)
paymo docs                  # Built-in documentation viewer
```

### List Flag Conventions
- `--limit, -l`: Number of results
- `--active, -a`: Active items only
- `--sort, -s`: Sort field
- `--date`: Date filter
- `--client`: Client filter
- `--user`: User filter
- `--project, -p`: Project filter

## Output Formatting

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
      "project": {"id": 456, "name": "paymo-cli"},
      "task": {"id": 789, "name": "Development"},
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
