# AI Agent Guidelines

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

## Documentation Maintenance

- **Maintain Roadmap**: Always update `ROADMAP.md` when completing a task or planning a new feature. Ensure the implementation status (legend) is accurate and reflects the current state of the project.

## Directives for Development

### Source Control (Jujutsu VCS)

- **Use Jujutsu**: This project uses Jujutsu VCS (<https://docs.jj-vcs.dev/latest/>). All source control operations MUST use the `jj` command instead of `git`.
- **Status and Diff**: Use `jj st` to check the status of the working copy and `jj diff` to review changes.
- **Committing Changes**: When asked to commit, use `jj commit -m "your message"`. This sets the description of the current change and automatically switches to a new, empty revision.
- **Describing Changes**: Use `jj describe -m "your message"` if you only need to update the description of the current revision without creating a new one.
- **New Revisions**: Use `jj new` to snapshot the current state and provide isolation for a new batch of changes.
- **History**: Use `jj log -n 5` to review recent history and maintain consistent description styles.
- **Automatic Tracking**: Note that Jujutsu automatically tracks changes in the working copy; explicit "adding" of files is generally not required unless they are new and ignored.
