# Implementation Plan: Migrate CLI to spf13/cobra

## Phase 1: Preparation [checkpoint: ]

- [ ] Task: Research existing CLI implementation and flag usage
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Preparation' (Protocol in workflow.md)

## Phase 2: Dependency & Basic Structure [checkpoint: ]

- [ ] Task: Add `spf13/cobra` and `spf13/viper` as dependencies
    - [ ] Run `go get github.com/spf13/cobra`
    - [ ] Run `go get github.com/spf13/viper`
- [ ] Task: Create basic Cobra command structure in `cmd/sudoku/`
    - [ ] Write Tests: Verify root command initialization
    - [ ] Implement: Create `root.go` and initialize `RootCmd`
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Dependency & Basic Structure' (Protocol in workflow.md)

## Phase 3: Core Feature Migration [checkpoint: ]

- [ ] Task: Migrate 'solve' functionality to a Cobra subcommand or root flags
    - [ ] Write Tests: Verify solve command with existing test puzzles
    - [ ] Implement: Port solving logic to Cobra command
- [ ] Task: Migrate 'benchmark' functionality to a 'bench' subcommand
    - [ ] Write Tests: Verify bench command executes and reports correctly
    - [ ] Implement: Port benchmarking logic to 'bench' subcommand
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Core Feature Migration' (Protocol in workflow.md)

## Phase 4: Finalization & Cleanup [checkpoint: ]

- [ ] Task: Refactor `main.go` to use the new Cobra entry point
- [ ] Task: Remove old flag handling logic
- [ ] Task: Manually verify all CLI commands and help text
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Finalization & Cleanup' (Protocol in workflow.md)
