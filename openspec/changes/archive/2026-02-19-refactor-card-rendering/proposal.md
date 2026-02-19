## Why

The windfury visual effect (`drawWindfury`) lives inside `CardRenderer` in `ui/widget/card.go`, tightly coupled to the card widget. The `ui/effect/` package already provides an `Effect` interface with `DrawBehind`/`DrawFront` passes — exactly what windfury needs. Moving windfury there improves cohesion and makes the card renderer thinner.

## What Changes

- Extract `CardRenderer.drawWindfury()` from `ui/widget/card.go` into a new `Windfury` effect struct in `ui/effect/`
- The new effect implements `effect.Effect` with `DrawBehind` (back pass) and `DrawFront` (front pass)
- Windfury is a persistent keyword effect (never "done"), unlike transient combat effects (flash, shake)
- Add `KindWindfury` to `effect.Kind` enum
- Card renderer delegates windfury drawing to the effect instead of calling its own method
- No visual change — the rendering output is identical

## Capabilities

### New Capabilities

### Modified Capabilities
- `card-rendering`: keyword effect drawing delegates to `ui/effect/` package instead of inline methods

## Impact

- `ui/widget/card.go` — remove `drawWindfury` method, add call to windfury effect's draw methods
- `ui/effect/` — new `windfury.go` file with `Windfury` struct implementing `Effect`
- `ui/effect/effect.go` — add `KindWindfury` constant
