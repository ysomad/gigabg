# Design: Fix json/v2 imports

## Context

The Makefile already sets `export GOEXPERIMENT=jsonv2`, so `make` targets work. However, gopls and direct `go build` outside of `make` don't inherit Makefile exports, causing IDE diagnostic errors.

## Goals / Non-Goals

**Goals:**
- Make `encoding/json/v2` resolve for all Go tooling (gopls, go build, go test)

**Non-Goals:**
- Changing any application code
- Removing the existing Makefile export (it's still useful for CI)

## Decision: Use `go env -w`

Run `go env -w GOEXPERIMENT=jsonv2` to persist the setting in the Go env file (`~/Library/Application Support/go/env`). This is the standard Go mechanism for persistent environment overrides and is read by all Go tools including gopls.

**Why not `.envrc` or IDE settings?** Those are tool-specific. `go env -w` is universal across all Go tooling.
