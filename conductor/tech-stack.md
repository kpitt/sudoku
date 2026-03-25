# Technology Stack

## Core Language
- **Go (Go 1.24):** The project is built using Go, leveraging its performance and strong concurrency primitives.

## CLI & Configuration
- **spf13/cobra:** A powerful library for creating CLI applications.
- **spf13/viper:** A complete configuration solution for Go applications, supporting multiple formats and environments.

## Key Libraries
- **github.com/fatih/color:** Used for terminal color output.
- **github.com/mattn/go-isatty:** Used to detect if the output is a terminal (for color handling).

## Architecture
- **CLI Tool:** A standalone command-line application.
- **Internal Packages:** Organized logic for bitsets, puzzles, sets, and solvers.
