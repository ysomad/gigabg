## Context

`CardRenderer` has four public draw methods: `DrawHoverCard`, `DrawHandCard`, `DrawShopCard`, `DrawMinion`. All take `api.Card` (the wire type). `DrawHoverCard` internally looks up the template via `r.Cards.ByTemplateID(c.Template)` to get name, description, and tribes — data that only exists on `game.CardTemplate`.

The full-detail card view is the canonical card representation. It should take template data directly, not a wire snapshot.

## Goals / Non-Goals

**Goals:**
- Rename `DrawHoverCard` → `DrawCard` to reflect it's the canonical full-detail card render
- Change `DrawCard` to accept `game.CardTemplate` instead of `api.Card`
- Move template lookup responsibility to the caller

**Non-Goals:**
- Changing compact card renderers (`DrawHandCard`, `DrawShopCard`, `DrawMinion`) — they show runtime state and correctly take `api.Card`
- Adding effect or aura rendering to the card
- Changing the visual layout of the full card

## Decisions

**1. `DrawCard` takes `game.CardTemplate` (not a new view struct)**

The card renders pure template data. `game.CardTemplate` already exposes everything needed: `Name()`, `Description()`, `Tier()`, `Tribes()`, `Attack()`, `Health()`, `Keywords()`. No wrapper type needed.

Alternative considered: a `CardView` struct combining `api.Card` + template data. Rejected — unnecessary indirection for now. Can be introduced later when runtime-modified effects need display.

**2. Internal helpers renamed: `drawHoverMinion` → `drawCardMinion`, `drawHoverSpell` → `drawCardSpell`**

Signature changes to `game.CardTemplate`. They read fields directly from the template parameter instead of going through `cardInfo()`.

**3. `isSpell(api.Card)` not needed for `DrawCard` path**

`DrawCard` receives a `CardTemplate` which has `Kind()` directly. The `isSpell` helper (which does a catalog lookup from `api.Card.Template`) is only used by the compact renderers.

## Risks / Trade-offs

**[Caller must handle nil template]** → The caller does the lookup and must guard against nil. This is a single call site (`recruit.go`) with a simple nil check.
