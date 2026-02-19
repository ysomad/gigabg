## Why

`DrawHoverCard` is named after a UI interaction (hovering), not what it actually is: the canonical full-detail card render showing all template data. The method also takes `api.Card` (a wire type) and internally looks up the template, when it should take `game.CardTemplate` directly since it only renders template data.

## What Changes

- **BREAKING**: Rename `CardRenderer.DrawHoverCard` to `CardRenderer.DrawCard`
- Change `DrawCard` signature from `api.Card` to `game.CardTemplate`
- Rename internal helpers `drawHoverMinion` / `drawHoverSpell` to `drawCardMinion` / `drawCardSpell`, update their signatures to take `game.CardTemplate`
- Move template lookup (`catalog.ByTemplateID`) to the caller in `recruit.go`
- Remove `cardInfo()` usage from `DrawCard` path — read name, description, tribes directly from the template parameter

## Capabilities

### New Capabilities

- `card-rendering`: Card rendering API for the `CardRenderer` widget — method names, signatures, and responsibilities for each card view (full-detail, hand, shop, board)

### Modified Capabilities

_None — no existing specs._

## Impact

- `ui/widget/card.go`: Rename methods, change signatures, update internal helpers
- `ui/scene/recruit.go`: Update call site to look up template and pass `game.CardTemplate` instead of `api.Card`
