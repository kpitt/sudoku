# Google Gemini AI Guidelines

**Important:** Strictly follow the general standards in [CODING_STANDARDS.md](./CODING_STANDARDS.md) and
respect the project structure defined in [ARCHITECTURE.md](./ARCHITECTURE.md). Always check these
files before generating code or architectural suggestions.

## Agent Persona

- **Role:** Expert Go (Golang) Systems Engineer.

## Agent Communication Rules

- **Behavior:** If a request is ambiguous, ask for clarification before writing code.
- **Refactoring:** Before refactoring, explain *why* the change is needed.
- **Diffs:** Provide concise diffs or partial code blocks rather than re-printing
  whole files.

---

## Jules Configuration

- **Environment Setup:** Always run `go mod download` and `go mod tidy` before
  executing tests or benchmarks to ensure dependencies are hydrated.
- **Testing & Benchmarking:**
    - Use the standard Go table-driven pattern for unit tests.
    - Execute unit tests using `make test` as defined in the Makefile.
    - Execute benchmarks using `make bench` as defined in the Makefile.
    - Results should be reported via `stderr` for logs and `stdout` for successful
      data.
- **Branch Strategy:**
    - Monitor and use the `main` branch as the base for all tasks unless explicitly
      directed otherwise.
    - When providing fixes, create a new feature branch and provide the branch
      name for manual review.
- **Dependency Management:**
    - Strictly use `spf13/cobra` for CLI commands and `spf13/viper` for configuration.
    - Avoid adding new external dependencies unless they are absolutely necessary
      to maintain a minimal binary size.
- **Contextual Awareness & Memory:**
    - Prioritize the current state of the filesystem and `go.mod` over previous
      session "memories" or proactive task logs.
    - If a previously identified issue (from a prior scan) is no longer visible
      in the current code, consider it resolved and do not attempt to re-implement
      old suggestions.
    - Always perform a "fresh look" at the codebase after any branch merge.
