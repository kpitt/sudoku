# Coding Standards: Sudoku CLI Application

## Project Context & Philosophy

- **Project Type:** High-performance Command Line Interface (CLI).
- **Core Philosophy:** Unix philosophy (do one thing well), fast startup times,
  and minimal binary size.

## Go Language Standards

- **Idiomatic Go:** Follow "Effective Go" and `golangci-lint` recommendations.
- **Error Handling:** Errors are values. Never use `panic()` for flow control.
  Wrap errors with context: `fmt.Errorf("doing thing: %w", err)`.
- **Structs:** Prefer passing pointers to large structs; use value receivers for
  small, immutable types.
- **Generics:** Use generics where appropriate to reduce code duplication.
- **Language Features:** Use the latest language features up to and including Go 1.25.
- **Maintenance:** Always run `go mod tidy` after adding dependencies.

## Tech Stack & Frameworks

- **CLI Engine:** Use `spf13/cobra`. All new commands must be generated using the
  Cobra pattern.
- **Configuration:** Use `spf13/viper`. Bind all flags to config keys in the `init()`
  function of commands.

## CLI Design Principles

- **Stdout vs Stderr:** Normal output goes to `stdout`; logs, progress bars, and
  errors go to `stderr`.
- **Piping:** Ensure the utility supports piped input (`stdin`) and output where
  applicable.
- **Exit Codes:** Use `0` for success, `1` for general errors, and specific codes
  for logic-specific failures.
- **Arguments vs Flags:** Use arguments for required data (files/IDs) and flags
  for options.

## Cross-Platform & Terminal Compatibility

- **Pathing:** Use `filepath.Join()` for all file paths; never hardcode `/` or `\`.
- **Line Endings:** Use `\n` for internal logic; be mindful of `\r\n` when reading
  Windows files.
- **Color & Styling:** Use `github.com/fatih/color` to ensure ANSI escape codes
  are handled correctly. Respect `NO_COLOR` environment variables.
- **Windows Support:** For Windows-specific features, check `runtime.GOOS == "windows"`.
  Use build tags (`//go:build windows`) for platform-specific files.
- **Platform-Specific Files:** Use build tags (e.g., `//go:build windows`) to
  separate platform-specific logic into `_windows.go` and `_unix.go` files.

## Testing & Documentation

- **Testing:** Use the standard Go table-driven pattern for unit tests.
- **Mocking:** Use interfaces to mock external I/O (filesystem, network) for unit
  tests.
- **Documentation:** Every exported symbol must have a comment starting with the
  symbol name.

## Development Workflow

### Prerequisites

- **Go:** Version 1.24 or higher.
- **Make:** For running build and test commands.

### Building

To build the executable:

```bash
make build
```

This will create the `bin/sudoku` binary.

### Running

The solver reads from `stdin`. You can pipe a puzzle file to it:

```bash
# Run with a sample puzzle
cat test/expert1.txt | bin/sudoku
```

Or run interactively:

```bash
bin/sudoku
# Paste your puzzle grid here...
```

### Testing

Run the test suite:

```bash
make test
```

Run benchmarks:

```bash
make bench
```

### Development Makefile Targets

The `Makefile` provides several useful targets for development:

- `make fmt`: Format code using `go fmt`.
- `make lint`: Run linters (requires `golangci-lint`).
- `make vet`: Run `go vet`.
- `make test-coverage`: Generate and view a test coverage report.
- `make dev`: Runs format, vet, and tests in one go.
