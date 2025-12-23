# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **danger-go**, a Go implementation of the popular Danger tool that runs automation rules during PR reviews. It's a wrapper around Danger JS that allows writing Dangerfiles in Go instead of JavaScript.

The project consists of:

- A Go library for writing Danger rules (`api.go`, `types.go`)
- A command-line tool (`cmd/danger-go/`) that wraps Danger JS
- Type definitions for various platforms (GitHub, GitLab) in `danger-js/` directory
- Platform-specific types in separate files (`types_*.go`)

## Architecture

### Core Components

1. **API Layer** (`api.go`): Main public API with the `T` struct providing methods:
   - `Message()` - Add informational messages
   - `Warn()` - Add warnings that don't fail the build
   - `Fail()` - Add failures that fail the build
   - `Markdown()` - Add raw markdown to comments

2. **Types** (`types.go`): Core data structures like `Results`, `Violation`, `GitHubResults`

3. **Danger JS Bridge** (`danger-js/`): Integration layer that:
   - Calls the `danger` (JS) binary to get DSL data
   - Processes commands (`ci`, `local`, `pr`) by wrapping Danger JS
   - Contains platform-specific type definitions

4. **CLI Tool** (`cmd/danger-go/`): Command-line interface supporting:
   - `ci` - Run on CI/CD
   - `local` - Run locally for git hooks
   - `pr` - Test against existing GitHub PR
   - `runner` - Internal command for processing DSL via STDIN

## Development Commands

### Building and Testing

```bash
# Run tests
go test -v ./...

# Build the CLI tool
go build -o danger-go cmd/danger-go/main.go

# Install the CLI tool globally
go install github.com/danger/golang/cmd/danger-go@latest
```

### Running Danger Locally

```bash
# Install dependencies first
npm install -g danger
go install github.com/danger/golang/cmd/danger-go@latest

# Run danger in CI mode (from build/ci directory)
cd build/ci && danger-go ci

# Run locally for testing
danger-go local

# Test against a specific PR
danger-go pr https://github.com/owner/repo/pull/123
```

### Development Workflow

The project follows standard Go conventions:

- Use `go fmt` for formatting
- Run `go vet` for static analysis
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write table-driven tests where appropriate
- Use conventional commit messages

## Dangerfile Structure

Dangerfiles are Go programs that must:

1. Be in a separate directory (e.g., `build/ci/`)
2. Have a `Run(d *danger.T, pr danger.DSL)` function
3. Import `github.com/danger/golang`
4. Have their own go.mod file

Example Dangerfile setup:

```bash
mkdir build/ci
cd build/ci
go mod init dangerfile
go get github.com/danger/golang
```

## Key Dependencies

- **Go 1.24+** required
- **Danger JS** must be installed globally (`npm install -g danger`)
- **github.com/stretchr/testify** for testing

## CI/CD Integration

The project uses GitHub Actions (`.github/workflows/test.yml`) which:

- Installs Go 1.24+
- Installs Node.js and Danger JS
- Runs both Go tests and danger-go CI checks
- Requires `GITHUB_TOKEN` for GitHub API access

## Testing

- Use `go test -v ./...` to run all tests
- Tests are in `*_test.go` files
- Internal tests in `api_internal_test.go` test unexported functions
- Follow table-driven test patterns where applicable
