<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Output Formatting

Multiple output formats for different use cases.

## Quick Start

```bash
# Default console output
gitprovidersync sync

# JSON for automation
gitprovidersync sync --format=json

# Plain text
gitprovidersync sync --format=plain

# Errors only
gitprovidersync --quiet sync
```

## Output Streams

- **stdout**: Results and formatted output
- **stderr**: Errors and debug logs

## Output Formats

### Default Format (Console for TTY)

Colored output with icons for terminals:

```bash
gitprovidersync sync --dry-run
# or
gitprovidersync sync --format=default
```

### Plain Format (Default for Pipes)

Text output without colors:

```bash
gitprovidersync sync --dry-run --format=plain
```

Auto-selected when output is piped.

### JSON Format

Structured output for automation:

```bash
gitprovidersync sync --dry-run --format=json | jq '.type'
```

Newline-delimited JSON events.

- Ideal for monitoring and automation

## Verbosity Levels

Control the amount of information displayed:

### Quiet Mode

Completely silent except for errors:

```bash
gitprovidersync --quiet sync --dry-run
```

### Brief Mode (Default)

Standard output with progress and results:

```bash
gitprovidersync sync --dry-run
```

### Verbose Mode

Additional information about operations:

```bash
gitprovidersync --verbose sync --dry-run
```

### Debug Mode

Detailed debug logs to stderr:

```bash
gitprovidersync --log-level=debug sync --dry-run
```

### Trace Mode

Maximum verbosity for troubleshooting:

```bash
gitprovidersync --log-level=trace sync --dry-run
```

## Color Control

Control when colors are used in output:

- `--color=auto` (default): Use colors when outputting to a TTY
- `--color=always`: Always use colors
- `--color=never`: Never use colors
- Environment variable `NO_COLOR`: Disables colors when set

## Usage Examples

### Separate Streams

```bash
# Capture results to file, errors to screen
gitprovidersync sync > results.txt

# Capture errors to file, results to screen  
gitprovidersync sync 2> errors.log

# Capture both separately
gitprovidersync sync > results.txt 2> errors.log
```

### CI/CD Pipeline

```bash
# JSON output for parsing
gitprovidersync sync --format=json | process-results

# Plain output for logs
gitprovidersync sync --format=plain >> sync.log

# Quiet mode for cron jobs
gitprovidersync --quiet sync
```

### Debugging

```bash
# Debug logs to stderr, results to stdout
gitprovidersync --log-level=debug sync 2> debug.log

# Everything to a file for analysis
gitprovidersync --log-level=trace sync &> full-output.log
```

## Notes

- Output goes to stdout, logs to stderr
- Format auto-detected from TTY presence
- Colors respect NO_COLOR environment variable
