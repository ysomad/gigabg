## Why

The card hover tooltip currently takes over the entire screen — it renders a large card centered with a dim overlay, which obscures the game state. This is disruptive for quick card inspection during the recruit phase. Additionally, the tooltip appears instantly on hover with no delay, making it trigger accidentally when moving the mouse across cards.

## What Changes

- Reposition the hover tooltip to appear to the right of the hovered card instead of centered on screen
- Remove the full-screen dim overlay
- Add a hover delay so the tooltip only appears after the cursor stays on a card for a short duration
- Reset the delay timer when the cursor moves to a different card or leaves all cards

## Capabilities

### New Capabilities
- `hover-tooltip`: Card hover tooltip positioning (anchored to hovered card rect) and display delay behavior

### Modified Capabilities

## Impact

- `ui/scene/recruit.go` — `recruitPhase` struct gains a hover timer; `updateHover` tracks elapsed time; `drawHoverTooltip` computes tooltip rect relative to `hoverRect` instead of screen center
- `ui/scene/shop.go` — no changes expected (hover detection returns the same data)
- `ui/widget/card.go` — no changes expected (`DrawCard` signature and rendering unchanged)
