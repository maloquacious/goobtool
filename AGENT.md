# AGENT.md

## Purpose
This is the repository for `goobtool`, a sandbox testing our implementation of the Goobergine reference implementation.

The tool implements a command line interface for a web server along with commands to administer the application.

## Project Structure
    project-root/
    ├── cmd/                 # Command line applications go here
    │   └── app/             # CLI for admin and server
    │
    ├── dist/                # Build artifacts, one directory per deploy target
    │   ├── linux/           # Linux (production target)
    │   └── local/           # Local development
    │
    ├── docs/                # Documentation (Markdown, Diagrams, etc.)
    │
    ├── testdata/            # Data for testing the application
    │
    ├── tools/               # Development scripts and tools
    │   └── ... (dev tools, bash scripts, etc.)
    │
    ├── .gitattributes
    ├── .gitignore
    ├── go.mod
    ├── go.work              # Development may use local repositories
    ├── LICENSE
    ├── README.md
    └── ... (CI/CD configs, etc.)

## Commands
* CLI command:
  * Build CLI: `go build -o dist/local/app ./cmd/app`
  * Version info: `dist/local/app version`
  * Tests: `go test ./...`
  * Format code: `go fmt ./...`
  * Build for Linux: get version then `GOOS=linux GOARCH=amd64 go build -o dist/linux/app-${VERSION}`

## Code Style
- Standard Go formatting using `gofmt`
- Imports organized by stdlib first, then external packages
- Error handling: return errors to caller, log.Fatal only in main
- Function comments use Go standard format `// FunctionName does X`
- Variable naming follows camelCase
- File structure follows standard Go package conventions
- Use `_e`, `_i`, `_t` when declaring types
  - Enums: use `type Enum_e int`
  - Interface: use `type Interface_i interface {}`
  - Struct: `type Struct_t struct {}` naming

## Bash Scripts
- Always use `${VARIABLE}` with curly braces for all variables
- Always quote variable references: "${VARIABLE}"
- Use `set -e` for early exit on errors
- Include descriptive echo statements with emoji for visual feedback
- Test endpoints in sequence with explicit validation
- Exit with error code on test failures
- Use curl with proper headers and jq for parsing responses
