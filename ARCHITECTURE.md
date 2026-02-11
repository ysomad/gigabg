# Architecture

## Overview

Two binaries: a WebSocket game server with PostgreSQL for game results, and an Ebiten desktop client.

- **Game server** (`cmd/gameserver/`) — WebSocket server, lobbies in memory, writes game results to PG
- **Client** (`cmd/client/`) — Ebiten desktop client

### Player Identity

No auth. Players choose a player ID (string) when connecting to a lobby. This ID is visible to other players.

## Directory Tree

```
gigabg/
├── internal/
│   ├── game/                     # Pure domain logic (no I/O)
│   │   ├── cards/                # Card template definitions
│   │   └── *.go                 # board, combat, minion, player, shop, etc.
│   │
│   ├── lobby/                    # Game orchestration
│   │   └── lobby.go             # Phase management, pairing, combat execution
│   │
│   ├── pkg/errors/               # Constant error type
│   │
│   ├── gameserver/               # Game server (WebSocket)
│   │   └── server.go            # WS handler, game loop, message dispatch
│   │
│   ├── api/                      # API types
│   │   └── api.go               # WS messages, game state, card types
│   │
│   ├── client/                   # WebSocket client
│   │   └── client.go
│   │
│   └── ui/                       # Ebiten UI
│       ├── app.go               # Root ebiten.Game, owns SceneManager + font
│       ├── scene.go             # Scene interface + SceneManager
│       ├── layout.go            # Rect, CalcGameLayout, CalcCombatLayout
│       ├── draw.go              # DrawText, EaseOut, ColorBackground
│       ├── scene/               # Scene implementations
│       │   ├── menu.go          # Player ID + lobby ID entry
│       │   ├── game.go          # Recruit phase + combat display
│       │   └── combat.go        # CombatAnimator (animation replay)
│       └── widget/              # Reusable UI components
│           ├── button.go        # Clickable button
│           ├── textinput.go     # Text input field
│           └── card.go          # CardRenderer: hand/shop/board/combat cards
│
├── migrations/                   # SQL migration files (goose)
│
├── cmd/
│   ├── client/main.go           # ui.NewApp(), scene wiring, ebiten.RunGame()
│   ├── gameserver/main.go       # Loads cards, starts WS server
│   └── web/main.go              # WASM web server
│
└── web/                          # WASM assets
```

## Package Dependency Graph

All paths under `internal/`:
```
pkg/errors/     (leaf)
game/           → pkg/errors/
game/cards/     → game/
lobby/          → game/, pkg/errors/
api/            → game/
gameserver/     → api/, game/, lobby/, pkg/errors/
client/         → api/, game/
ui/             → ebiten
ui/widget/      → ui/, api/, game/cards/, ebiten
ui/scene/       → ui/, ui/widget/, client/, api/, game/, game/cards/, ebiten
cmd/client/     → ui/, ui/scene/, client/, game/cards/
cmd/gameserver/ → gameserver/, game/cards/
```

No circular dependencies.

## Key Design Decisions

- **No auth** — players identify by self-chosen player ID, no JWT/sessions
- **`internal/` for all packages** — single binary consumers in `cmd/`, everything else is internal
- **No interfaces unless necessary** — concrete types everywhere; interfaces only for testing or breaking circular deps
- **Scene interface is needed** — breaks circular dependency between `ui/` and `ui/scene/`
- **Widgets are concrete structs** — no layout engine; Ebiten immediate-mode drawing
- **Zone-based layout** — `Rect` type + `CalcGameLayout(w,h)` computes positions from screen size, no hardcoded base coordinates
- **pgx over database/sql** — native PG protocol, connection pooling, type safety
- **goose for migrations** — pure SQL migrations, simple Go API

## Dependencies

| Library | Purpose |
|---------|---------|
| `jackc/pgx/v5` | PostgreSQL driver + connection pool |
| `pressly/goose/v3` | SQL migrations |
| `coder/websocket` | WebSocket client/server |
| `hajimehoshi/ebiten/v2` | Game engine |
