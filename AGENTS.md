# Agent Development Guidelines

## Build/Lint/Test Commands
- Build: `go build -o cerebras-code-monitor cmd/main.go`
- Install dependencies: `go mod tidy`
- Run all tests: `go test ./...`
- Run single test: `go test -run TestName ./path/to/package`
- Lint: `golangci-lint run`
- Format: `go fmt ./...`

## Code Style Guidelines
- Use descriptive variable names in camelCase
- Function names should be clear and follow camelCase
- All structs and functions should have comments
- Error handling: check errors immediately after they occur
- Imports: group standard library imports separately from third-party
- Use Cobra for CLI commands and Viper for configuration
- Follow existing patterns in cmd/ and internal/cmd/ directories
- Use conventional commits for git commit messages (e.g. feat:, docs:, fix:, etc.)

## Testing
- Add tests for new functionality in *_test.go files
- Use table-driven tests when appropriate
- Mock external API calls in tests
- Verify error cases are handled

## Environment Variables
- CEREBRAS_SESSION_TOKEN: session authentication token
- CEREBRAS_API_KEY: API key authentication (alternative method)

## File Structure
- cmd/main.go: entry point
- internal/cmd/: command implementations
- internal/config/: XDG configuration handling
- config.yaml: default configuration