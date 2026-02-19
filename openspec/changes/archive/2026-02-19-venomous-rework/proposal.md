## Why

The game has two overlapping poison keywords — Poisonous (permanent, kills all damaged minions) and Venomous (one-shot, kills the first damaged minion then is removed). No card templates use Poisonous, making it dead code. We should remove it and keep only Venomous as the single poison mechanic. Additionally, when Venomous is consumed (the minion kills its target), the UI should visually reflect the removal so players see the venom was spent.

## What Changes

- **BREAKING**: Remove `KeywordPoisonous` constant and all references (String, Description, combat logic, UI rendering, tests)
- Remove unused `IsPoisoned()` method from Minion
- Simplify `applyPoison` to only handle Venomous logic (remove Poisonous branch)
- Update `DeathReasonPoison` comment to reference only Venomous
- Update hit-indicator spec and combat board UI to check only `KeywordVenomous` for poison hit variant
- Remove `drawPoisonous` card renderer method
- Existing `RemoveKeywordEvent` + `VenomBreak` effect in `combatboard.go` already handles the UI animation for venom consumption — no new UI work needed

## Capabilities

### New Capabilities

_(none — this is a simplification, not a new feature)_

### Modified Capabilities

- `hit-indicator`: Remove Poisonous keyword check; poison hit variant triggers only on Venomous

## Impact

- `game/keyword.go` — remove `KeywordPoisonous` constant, update String/Description
- `game/keyword_test.go` — remove Poisonous from test cases
- `game/minion.go` — remove unused `IsPoisoned()` method
- `game/combat.go` — simplify `applyPoison` to Venomous-only
- `game/combatevent.go` — update `DeathReasonPoison` doc comment
- `game/catalog/` — no changes needed (no cards use Poisonous)
- `ui/widget/card.go` — remove `drawPoisonous` method and its call site
- `ui/scene/combatboard.go` — remove Poisonous check in hit indicator selection
- `openspec/specs/hit-indicator/` — update spec to remove Poisonous references
