## Context

The hover tooltip in the recruit phase currently renders a 2.5x card centered on screen with a full-screen dim overlay (`drawHoverTooltip` in `ui/scene/recruit.go`). Hover is detected each frame in `updateHover` and stored as `hoverCard *api.Card` + `hoverRect ui.Rect`. There is no delay — the tooltip appears instantly.

All coordinates use the base coordinate system (1280x720). The tooltip must work with shop cards (top zone) and board cards (middle zone).

## Goals / Non-Goals

**Goals:**
- Position tooltip adjacent to the hovered card (right side preferred, left as fallback)
- Add a hover delay timer so the tooltip only appears after sustained hover
- Remove the dim overlay

**Non-Goals:**
- Changing `DrawCard` rendering or signature
- Adding hover tooltips to hand cards or discover cards
- Fade-in/fade-out animations for the tooltip
- Making the delay configurable at runtime (hardcoded constant is fine)

## Decisions

### 1. Tooltip sizing: 2.0x card dimensions (down from 2.5x)

The tooltip appears inline next to a card rather than as a screen-center modal. 2.5x is too large — it would overlap neighboring cards and potentially overflow screen bounds. 2.0x provides readable detail while fitting comfortably beside the hovered card.

**Alternative**: Keep 2.5x. Rejected because at 2.5x the tooltip width alone is 25% of BaseWidth, plus the card itself and gap — too likely to overflow horizontally.

### 2. Positioning: right-of-card with left fallback and vertical clamping

Compute tooltip rect in base coordinates:
- **Right placement**: `tooltipX = hoverRect.X + hoverRect.W + gap`
- **Left fallback**: if right edge (`tooltipX + tooltipW`) exceeds `BaseWidth`, place left: `tooltipX = hoverRect.X - gap - tooltipW`
- **Vertical center**: `tooltipY = hoverRect.Y + hoverRect.H/2 - tooltipH/2`
- **Vertical clamp**: clamp `tooltipY` to `[0, BaseHeight - tooltipH]`

Gap: `lay.Gap` (same gap used between cards).

**Alternative**: Always place tooltip at a fixed screen position (e.g. right edge). Rejected because anchoring to the hovered card provides better spatial context.

### 3. Hover state: track previous card identity + elapsed time

Add three fields to `recruitPhase`:
- `hoverElapsed float64` — seconds accumulated while hovering the same card
- `hoverPrevCard *api.Card` — the card from the previous frame (compared by pointer or template+position identity)
- `hoverReady bool` — whether the delay has been met

Each frame in `updateHover`:
1. Detect the currently hovered card (existing logic, sets `hoverCard`/`hoverRect`)
2. If `hoverCard` is nil or differs from previous frame → reset `hoverElapsed` to 0, `hoverReady` to false
3. If same card → increment `hoverElapsed` by `1/TPS` (Ebiten's `1.0/60.0`)
4. If `hoverElapsed >= hoverDelay` → set `hoverReady` to true
5. Store current `hoverCard` as `hoverPrevCard`

Card identity comparison: compare the `hoverRect` (X, Y position) since the same card at the same visual position has the same rect. This avoids needing a unique card ID and handles the case where cards are reordered.

**Alternative**: Use `time.Now()` to track real wall-clock time. Rejected because Ebiten's fixed-tick `Update` loop already provides deterministic time steps — accumulating `1/TPS` per tick is simpler and consistent with game loop timing.

### 4. Hover delay constant: 400ms

`const _hoverDelay = 0.4` (seconds). This is long enough to prevent accidental triggers when sweeping across cards, but short enough that intentional hovers feel responsive.

### 5. drawHoverTooltip only draws when hoverReady is true

Replace the current `if r.hoverCard == nil` guard with `if !r.hoverReady || r.hoverCard == nil`. Remove the `FillScreen` dim call entirely.

## Risks / Trade-offs

- **Tooltip may overlap adjacent cards**: With 2.0x sizing and a gap offset, the tooltip can overlap neighboring cards in the same row. This is acceptable because the tooltip is drawn last (on top) and the user is focused on the hovered card. If it becomes a problem, tooltip size can be reduced further.
- **No fade-in**: The tooltip appears abruptly after the delay. This is simpler and avoids adding animation state. Can be added later if desired.
