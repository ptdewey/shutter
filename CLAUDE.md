# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shutter is a snapshot testing library for Go, inspired by Rust's [insta](https://github.com/mitsuhiko/insta) and Gleam's [birdie](https://github.com/giacomocavalieri/birdie). It provides functions for capturing test output as snapshots and tools for reviewing snapshot changes.

## Build Commands

```bash
# Build the TUI binary
just build

# Run tests with coverage
just test

# Clean snapshots and run tests
just clean-test

# Run the TUI review tool (after building)
just review

# Run CLI review tool
just cli
```

Direct Go commands:
```bash
go test ./...                                    # Run all tests
go test ./... -run TestName                      # Run specific test
go test -v ./internal/transform/...              # Run tests in specific package
cd ./cmd/shutter && go build -o shutter ./main.go  # Build TUI
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Public API (shutter.go)                                        │
│  Snap() | SnapMany() | SnapString() | SnapJSON()                │
├─────────────────────────────────────────────────────────────────┤
│  Options (scrubbers.go, ignore.go)                              │
│  Scrubbers: text transformation before snapshot                 │
│  IgnorePatterns: field removal (SnapJSON only)                  │
├─────────────────────────────────────────────────────────────────┤
│  Internal Modules                                               │
│  ├─ internal/snapshots/ - Core comparison logic                 │
│  ├─ internal/files/     - Snapshot file I/O (YAML headers)      │
│  ├─ internal/transform/ - JSON ignore pattern application       │
│  ├─ internal/diff/      - Histogram diff algorithm              │
│  ├─ internal/pretty/    - Formatting and display boxes          │
│  └─ internal/review/    - Review workflow logic                 │
├─────────────────────────────────────────────────────────────────┤
│  Review Tools                                                   │
│  ├─ cmd/shutter/ - TUI (Bubbletea) - separate go.mod            │
│  └─ cmd/cli/     - CLI review tool                              │
└─────────────────────────────────────────────────────────────────┘
```

**Data flow:** Test Value → Pretty format (utter) → Ignore Patterns → Scrubbers → Snapshot file

**Snapshot storage:** `__snapshots__/` directories contain YAML-header files with metadata (title, test_name, file_name, version) followed by `---` delimiter and content.

## Key Design Decisions

- **Option interface pattern**: `Scrubber` and `IgnorePattern` both implement `Option` for type-safe compile-time separation
- **IgnorePatterns only work with SnapJSON()** - using them with Snap/SnapMany/SnapString returns an error
- **TUI is a separate Go module** (cmd/shutter/) to keep Bubbletea dependencies optional
- **Execution order**: Ignore patterns run first, then scrubbers

## Module Structure

- Root module (`go.mod`): Main library - Go 1.23.12+
- TUI module (`cmd/shutter/go.mod`): Separate module with Bubbletea dependencies - Go 1.25.2
- `/editors/`: Tree-sitter grammar for snapshot format (Node.js/Rust/Python/Swift bindings)
