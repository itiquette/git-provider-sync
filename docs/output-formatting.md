# Output Formatting

Git Provider Sync provides flexible output formatting to support different use cases, from interactive terminal usage to CI/CD automation.

## Quick Start

```bash
# Default: Beautiful console output for humans
gitprovidersync sync

# For CI/CD pipelines and automation
gitprovidersync sync --format=json

# For log files and scripts
gitprovidersync sync --format=plain

# Silent operation (errors only)
gitprovidersync --quiet sync
```

## Output Streams

The tool follows Unix conventions for output streams:

- **stdout**: User-facing data, results, and formatted output
- **stderr**: Error messages and debug logs (when enabled)

This separation allows for clean pipeline usage and log redirection.

## Output Formats

### Console Format (Default for TTY)

Beautiful, colored output with icons for interactive terminal usage:

```bash
gitprovidersync sync --dry-run
```

Features:
- Colored output with emoji icons
- Progress indicators
- Hierarchical display of operations
- Detailed summary with next steps

### Plain Format (Default for Pipes)

Simple text output without colors, ideal for logs and non-interactive environments:

```bash
gitprovidersync sync --dry-run --format=plain
```

Features:
- No colors or special characters
- Clean text suitable for log files
- Structured but simple output
- Auto-selected when output is piped

### JSON Format

Structured JSON output for automation and programmatic consumption:

```bash
gitprovidersync sync --dry-run --format=json | jq '.type'
```

Features:
- Newline-delimited JSON events
- Machine-readable format
- Complete operation details
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

## Architecture

The output formatting system follows hexagonal architecture:

1. **Domain Layer** (`internal/domain/ports/sync_formatter.go`):
   - Defines the `SyncOutputFormatter` interface
   - Pure domain concept, no implementation details

2. **Adapter Layer** (`internal/adapters/cli/`):
   - `console_formatter.go`: TTY-friendly output
   - `plain_formatter.go`: Simple text output
   - `json_formatter.go`: Structured JSON
   - `quiet_formatter.go`: Minimal output
   - `formatter_factory.go`: Creates appropriate formatter

3. **Integration** (`cmd/synccmd/formatter_integration.go`):
   - Connects formatters to sync command
   - Handles formatter selection based on environment

## Implementation Notes

- Formatters write to stdout for user data
- Logger writes to stderr only in debug/verbose modes
- In normal operation, logger is suppressed to avoid mixed output
- Output format detection is automatic based on TTY presence
- Color support respects NO_COLOR environment variable