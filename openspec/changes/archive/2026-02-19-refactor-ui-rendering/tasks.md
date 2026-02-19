## 1. Hover delay state

- [x] 1.1 Add `hoverElapsed float64`, `hoverPrevRect ui.Rect`, and `hoverReady bool` fields to `recruitPhase` struct in `ui/scene/recruit.go`
- [x] 1.2 Add `const _hoverDelay = 0.4` (seconds) in `ui/scene/recruit.go`
- [x] 1.3 Update `updateHover` to accumulate `hoverElapsed` by `1.0/ebiten.TPS()` when the hovered card rect matches `hoverPrevRect`, reset to 0 when it changes or becomes nil, and set `hoverReady = true` when `hoverElapsed >= _hoverDelay`

## 2. Tooltip positioning

- [x] 2.1 Replace the centered-on-screen tooltip rect in `drawHoverTooltip` with right-of-card positioning: `tooltipX = hoverRect.X + hoverRect.W + lay.Gap`, vertically centered on the hovered card
- [x] 2.2 Add left-side fallback: if `tooltipX + tooltipW > BaseWidth`, place tooltip to the left of the hovered card
- [x] 2.3 Clamp `tooltipY` to `[0, BaseHeight - tooltipH]` to prevent vertical overflow
- [x] 2.4 Change tooltip size from 2.5x to 2.0x card dimensions

## 3. Remove dim overlay

- [x] 3.1 Remove the `ui.FillScreen(screen, res, color.RGBA{0, 0, 0, 120})` call from `drawHoverTooltip`
- [x] 3.2 Guard `drawHoverTooltip` with `if !r.hoverReady || r.hoverCard == nil` instead of just `r.hoverCard == nil`
