## 1. Effect type refactor

- [x] 1.1 Add `HitType` enum (`HitTypeDamage`, `HitTypePoison`) to `ui/effect/` package
- [x] 1.2 Rename `KindHitDamage` to `KindHitIndicator` in `ui/effect/effect.go`
- [x] 1.3 Rename `HitDamage` struct to `HitIndicator` in `ui/effect/hitdamage.go`, rename file to `hitindicator.go`
- [x] 1.4 Update `NewHitIndicator` constructor to accept `HitType` and set variant-specific colors (`HitTypeDamage`: gold bg RGBA 180,140,10; `HitTypePoison`: green bg RGBA 30,160,50)

## 2. Poison variant rendering

- [x] 2.1 Add `drawSkull` helper method on `HitIndicator` that draws a procedural skull icon (circle head, two eye sockets, jaw line) using `vector` shapes
- [x] 2.2 Branch `DrawFront` on `HitType`: damage variant draws "-X" text, poison variant calls `drawSkull`

## 3. Combat board integration

- [x] 3.1 Add `IsPoisoned()` to `game.Minion`, add `IsPoisoned` field to `game.DamageEvent`, set it in `dealDamage` from server
- [x] 3.2 Add `hitTypeFromDamage` helper, replace `effect.NewHitDamage(...)` with `effect.NewHitIndicator(hitType, ...)` in `applyDamage`

## 4. Cleanup

- [x] 4.1 Update all references to `KindHitDamage` across the codebase to `KindHitIndicator`
- [x] 4.2 Verify build compiles with `go build ./...`
- [x] 4.3 Run existing tests with `go test ./...`
