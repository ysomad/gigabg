# Fix json/v2 imports

## Why

The project uses `encoding/json/v2` and `encoding/json/jsontext` (Go 1.26 stdlib), but these packages are gated behind `GOEXPERIMENT=jsonv2`. Without that environment variable, the build fails across 5 files with "build constraints exclude all Go files" errors.

## What Changes

Set `GOEXPERIMENT=jsonv2` in the Go environment so the existing json/v2 imports compile without errors.

## Impact

- `api/api.go` — uses `jsontext.Value`, `json.Marshal`
- `client/client.go` — uses `json.Marshal`, `json.UnmarshalRead`
- `client/game.go` — uses `json.Marshal`, `json.Unmarshal`
- `server/server.go` — uses `json.UnmarshalRead`, `json.MarshalWrite`, `json.Marshal`, `json.Unmarshal`
- `ui/scene/combatboard.go` — uses `jsontext.Value`, `json.Unmarshal`

No code changes required — only environment/build configuration.
