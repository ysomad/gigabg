# Tasks

## 1. Set GOEXPERIMENT

- [x] 1.1 Run `go env -w GOEXPERIMENT=jsonv2` to persist the setting

## 2. Verify

- [x] 2.1 Confirm `go env GOEXPERIMENT` outputs `jsonv2`
- [x] 2.2 Confirm `go build ./...` succeeds without import errors
