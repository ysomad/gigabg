## 1. Effect package changes

- [x] 1.1 Add `KindWindfury` to the `Kind` enum in `ui/effect/effect.go`
- [x] 1.2 Create `ui/effect/windfury.go` with `Windfury` struct implementing `Effect` interface
- [x] 1.3 Add `var _ Effect = (*Windfury)(nil)` compile-time check
- [x] 1.4 Implement `NewWindfury()` constructor, `Kind()`, `Update()` (elapsed-based angle accumulation), `Progress()` (returns 0), `Modify()` (no-op)
- [x] 1.5 Implement `DrawBehind()` — back pass with `alpha * 0.05` (dimmer streaks behind ellipse)
- [x] 1.6 Implement `DrawFront()` — front pass with `alpha * 0.18` (brighter streaks on top of badges)
- [x] 1.7 Port orbit geometry from `CardRenderer.drawWindfury`: orbit radius `ry*1.05`, 39° inclination, 2 orbits, 2 streaks/orbit, 36 steps, wave modulation, taper, RGBA{175,175,178}

## 2. Card renderer integration

- [x] 2.1 Add `windfury *effect.Windfury` field to `CardRenderer`
- [x] 2.2 Update `drawEllipseBase` to call `windfury.DrawBehind()` instead of `r.drawWindfury(..., false)` when keyword is present (lazy-init the effect)
- [x] 2.3 Update `drawKeywordEffects` to call `windfury.DrawFront()` instead of `r.drawWindfury(..., true)` when keyword is present
- [x] 2.4 Add `windfury.Update(elapsed)` call to advance animation each frame (derive elapsed from tick delta or pass it through)
- [x] 2.5 Remove `CardRenderer.drawWindfury()` method from `ui/widget/card.go`

## 3. Verification

- [x] 3.1 Build compiles with no errors (`go build ./...`)
- [ ] 3.2 Visual check: windfury minion in shop renders identically (wind ribbons orbit, front/back pass visible)
- [ ] 3.3 Visual check: windfury minion on board renders identically during combat
