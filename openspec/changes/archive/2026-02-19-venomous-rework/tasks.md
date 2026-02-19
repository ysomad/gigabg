## 1. Remove KeywordPoisonous from game package

- [x] 1.1 Remove `KeywordPoisonous` from the const block in `game/keyword.go` (subsequent keywords shift down)
- [x] 1.2 Remove `KeywordPoisonous` cases from `String()` and `Description()` in `game/keyword.go`
- [x] 1.3 Remove unused `IsPoisoned()` method from `game/minion.go`
- [x] 1.4 Update `game/keyword_test.go` — remove `KeywordPoisonous` from all test cases, adjust expected values

## 2. Simplify combat poison logic

- [x] 2.1 Rename `applyPoison` → `applyVenom`, simplify to Venomous-only, rename `poisonKilled` → `venomKilled`
- [x] 2.2 Update `DeathReasonPoison` comment in `game/combatevent.go` to reference only Venomous

## 3. Remove Poisonous from UI

- [x] 3.1 Remove `drawPoisonous` method and its call site in `ui/widget/card.go`

## 4. Verify

- [x] 4.1 Run `go build ./...` to confirm no compile errors
- [x] 4.2 Run `go test ./game/...` to confirm all tests pass
