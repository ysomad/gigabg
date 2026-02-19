## Context

The game currently defines two poison keywords: `KeywordPoisonous` (permanent — kills every minion damaged) and `KeywordVenomous` (one-shot — kills the first minion damaged, then is removed). No card templates use `KeywordPoisonous`, making it dead code. The `IsPoisoned()` method on Minion is also unused. The combat engine (`applyPoison`) branches on both keywords, and the card renderer has a separate `drawPoisonous` method. The UI already handles Venomous removal via `RemoveKeywordEvent` → `VenomBreak` effect animation.

## Goals / Non-Goals

**Goals:**
- Remove `KeywordPoisonous` and all its references to eliminate dead code
- Remove unused `IsPoisoned()` method
- Simplify `applyPoison` to handle only Venomous
- Keep existing Venomous consume-on-kill behavior and UI animation unchanged

**Non-Goals:**
- Changing Venomous combat semantics (already works correctly)
- Adding new visual effects for venom consumption (VenomBreak effect already exists)
- Renaming `applyPoison` or `DeathReasonPoison` — these names are still accurate for Venomous

## Decisions

### 1. Remove KeywordPoisonous from iota block (shift values)

Remove the `KeywordPoisonous` line from the const block, letting subsequent keywords shift down by 1. `KeywordVenomous` moves from value 5 to 4.

**Why:** Client and server are compiled from the same source, and keyword values are not persisted to the database. Shifting is cleaner than leaving a dead gap in the bitmask. The `Keywords` bitmask (`uint16`) and `Keyword` (`uint8`) types remain unchanged.

**Alternative:** Leave a `_ // removed` placeholder to preserve numeric values. Rejected — adds confusion for no benefit since nothing stores these values externally.

### 2. Simplify applyPoison in-place

Remove the `isPoisonous` branch from `applyPoison`. The function becomes: check `HasKeyword(KeywordVenomous)` → kill target → remove keyword → emit `RemoveKeywordEvent`. No rename needed — "poison" is still accurate for venom mechanics.

**Why:** Minimal diff, function still makes sense semantically.

### 3. Keep DeathReasonPoison name

Update the comment from "killed by Poisonous or Venomous" to "killed by Venomous" but keep the constant name `DeathReasonPoison`. The name is still semantically correct (venom is a type of poison).

### 4. No UI changes beyond removing drawPoisonous

The `combatboard.go` already plays `VenomBreak` animation when processing `RemoveKeywordEvent` for `KeywordVenomous`. Removing the Poisonous code path from the hit indicator check is the only UI-side change needed.

## Risks / Trade-offs

- **Keyword iota shift changes wire values** → Mitigated: client and server always deploy together from the same binary. No persistent keyword storage exists.
- **Tests reference KeywordPoisonous** → Will fail at compile time, easy to find and fix.
