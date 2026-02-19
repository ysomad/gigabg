# Go Environment: GOEXPERIMENT=jsonv2

## Requirement: GOEXPERIMENT persisted in Go env

`GOEXPERIMENT=jsonv2` must be set in the Go environment file (`go env GOENV`) so that all Go tooling — including gopls, IDE integrations, and direct `go build` — can resolve `encoding/json/v2` and `encoding/json/jsontext`.

### Scenario: gopls resolves json/v2 imports

- **WHEN** a developer opens the project in an IDE with gopls
- **THEN** no "could not import encoding/json/v2" diagnostics appear
- **AND** no "build constraints exclude all Go files" errors appear

### Scenario: go build succeeds without Makefile

- **WHEN** a developer runs `go build ./...` directly (not via `make`)
- **THEN** the build succeeds without import errors
