### Requirement: DrawCard renders full-detail card from CardTemplate
`CardRenderer.DrawCard` SHALL accept a `game.CardTemplate` and render the canonical full-detail card view showing: tier badge, portrait, name, description, tribes, keywords, attack/health badges.

#### Scenario: Minion template rendered
- **WHEN** `DrawCard` is called with a minion `CardTemplate`
- **THEN** the card displays tier badge, portrait, name, description, keyword labels, attack badge, health badge, and tribe label

#### Scenario: Spell template rendered
- **WHEN** `DrawCard` is called with a spell `CardTemplate`
- **THEN** the card displays cost, "SPELL" label, name, and description

### Requirement: DrawCard does not depend on api.Card
`CardRenderer.DrawCard` SHALL NOT accept `api.Card`. All displayed data MUST come from the `game.CardTemplate` parameter directly.

#### Scenario: No wire type in signature
- **WHEN** a caller invokes `DrawCard`
- **THEN** the method signature accepts `game.CardTemplate`, not `api.Card`

### Requirement: Caller performs template lookup
The call site SHALL look up the `CardTemplate` from the catalog before calling `DrawCard`. If the template is nil, `DrawCard` SHALL NOT be called.

#### Scenario: Recruit phase hover tooltip
- **WHEN** a player hovers over a card during recruit phase
- **THEN** the scene looks up the template via `catalog.ByTemplateID` and calls `DrawCard` with the result
- **THEN** if the template lookup returns nil, no card is drawn
