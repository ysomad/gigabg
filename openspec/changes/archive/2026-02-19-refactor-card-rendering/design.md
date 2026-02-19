## Context

`CardRenderer.drawWindfury()` in `ui/widget/card.go` renders wind ribbons in two passes — a back pass (behind the ellipse body, called from `drawEllipseBase`) and a front pass (on top of badges, called from `drawKeywordEffects`). The `ui/effect/` package already defines an `Effect` interface with `DrawBehind`/`DrawFront` that maps directly to these two passes, but currently only holds transient combat effects (flash, shake, death fade, etc.).

## Goals / Non-Goals

**Goals:**
- Move windfury drawing logic into `ui/effect/Windfury` struct implementing `Effect`
- Keep the rendered output pixel-identical

**Non-Goals:**
- Moving other keyword effects (taunt, divine shield, reborn, venomous, stealth) — future work
- Changing windfury visual design
- Changing the combat animation system

## Decisions

### 1. Persistent vs transient effect

Existing effects are transient: they have a duration, `Update` counts down, and returns `true` when done. Windfury is persistent — it renders for as long as the keyword is active.

**Decision**: `Windfury.Update` never returns `true`. It increments an internal angle counter using `elapsed` to drive the rotation animation. The caller is responsible for removing the effect when the keyword is lost.

**Alternative considered**: Making `Update` return `true` and re-adding each frame — rejected because it defeats the purpose of the effect list and wastes allocations.

### 2. Animation timing: tick counter vs elapsed seconds

Current implementation uses `CardRenderer.Tick` (an integer frame counter) via `float64(r.Tick) * 0.04`. The effect package uses `elapsed float64` (seconds since last frame).

**Decision**: Use elapsed-based animation. Accumulate elapsed into an `angle float64` field: `e.angle += elapsed * 2.4` (equivalent to `tick * 0.04` at 60fps). This is frame-rate independent and consistent with the rest of the effect package.

### 3. Where the effect lives: card renderer vs effect list

Currently, `drawEllipseBase` and `drawKeywordEffects` both check `keywords.Has(KeywordWindfury)` and call the draw method. After extraction, these call sites delegate to a `Windfury` effect instance instead.

**Decision**: The `CardRenderer` holds an optional `*effect.Windfury` field. When rendering a card with the windfury keyword, it lazily creates the effect (or reuses the existing one). `drawEllipseBase` calls `windfury.DrawBehind(...)` and `drawKeywordEffects` calls `windfury.DrawFront(...)`. The card renderer calls `windfury.Update(elapsed)` each frame to advance the animation. This avoids coupling to the combat effect list — the card renderer manages keyword effects independently.

**Alternative considered**: Adding windfury to the per-minion `effect.List` in combat board — rejected because windfury rendering is needed in shop cards too, not just combat.

### 4. Effect interface compliance

`Windfury` implements `Effect` fully: `Kind` returns `KindWindfury`, `Modify` is a no-op, `Progress` returns 0 (never completes). This allows it to be used in effect lists if needed in the future, even though the card renderer manages it directly for now.

## Risks / Trade-offs

- **Persistent effect is new pattern** → Low risk. The `Effect` interface doesn't assume transience; `Update` returning `false` is valid. Adding a comment clarifies intent.
- **Frame-rate dependent equivalence** → The `elapsed * 2.4` rate produces the same visual at 60fps. At other frame rates the speed is identical (which is an improvement over the tick-based version that would spin faster at higher fps).
