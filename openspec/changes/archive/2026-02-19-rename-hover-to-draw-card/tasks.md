## 1. Rename and update CardRenderer methods

- [x] 1.1 Rename `DrawHoverCard` to `DrawCard`, change signature from `api.Card` to `game.CardTemplate` in `ui/widget/card.go`
- [x] 1.2 Rename `drawHoverMinion` to `drawCardMinion`, update signature to `game.CardTemplate`, read fields directly from template parameter
- [x] 1.3 Rename `drawHoverSpell` to `drawCardSpell`, update signature to `game.CardTemplate`, read fields directly from template parameter

## 2. Update call site

- [x] 2.1 In `ui/scene/recruit.go` `drawHoverTooltip`, look up template via `r.cr.Cards.ByTemplateID`, guard nil, call `DrawCard` with `game.CardTemplate`

## 3. Verify

- [x] 3.1 Run `go build ./...` to confirm no compile errors
