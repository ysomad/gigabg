# Go Style Guide

## Core Rules

- Never use pointers to interfaces
- Copy slices/maps at boundaries to prevent mutation of internal state
- Use `defer` for cleanup (files, locks, connections)
- Channels: size 0 (unbuffered) or 1 only
- Enums: start at `iota + 1`, not `iota`
- No `panic` in library code
- No mutable globals; use dependency injection
- No type embedding in public structs
- Avoid `init()`; prefer explicit initialization
- Call `os.Exit`/`log.Fatal` only in `main()`
- Use `var _ Interface = (*Implementation)(nil)` to check interface implementations

## Errors

Decision table:
| Need matching? | Message type | Use |
|----------------|--------------|-----|
| No | static | `errors.New` |
| No | dynamic | `fmt.Errorf` |
| Yes | static | `const ErrX errors.Error = "error"` |
| Yes | dynamic | custom error type |

Rules:
- Always handle errors
- Handle errors once: either wrap+return OR log+handle, never both
- Wrap with context: `fmt.Errorf("get user %q: %w", id, err)`
- Export error vars only if callers need to match them

## Goroutines

- Every goroutine must have predictable termination
- Always provide a way to wait: `sync.WaitGroup` or done channel
- Test for leaks with `go.uber.org/goleak`

## Naming

- Packages: lowercase, short, singular, no `util/common/shared`
- Functions: MixedCaps (tests may use `Test_X_Y`)
- Unexported globals: prefix with `_` (except `err` prefix for errors)
- Import aliases: only when package name differs from path or conflicts exist

## Code Style

- Reduce nesting: handle errors/edge cases first, return early
- Reduce variable scope: use `if err := fn(); err != nil`
- Omit unnecessary else when variable is set in both branches
- Use raw string literals to avoid escaping: `` `text` ``
- Structs: use field names, omit zero values, use `var s Struct` for zero structs
