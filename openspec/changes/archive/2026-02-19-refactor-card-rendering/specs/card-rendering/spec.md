## ADDED Requirements

### Requirement: Windfury effect struct in effect package
`effect.Windfury` SHALL implement the `effect.Effect` interface. It SHALL render two crossing orbits of wind ribbons around the minion ellipse, split into a behind-card pass (`DrawBehind`) and a front-of-card pass (`DrawFront`).

#### Scenario: Windfury draws behind and in front
- **WHEN** a minion has the Windfury keyword
- **THEN** `DrawBehind` renders dimmer wind streaks behind the ellipse body
- **THEN** `DrawFront` renders brighter wind streaks on top of the badges

#### Scenario: Windfury is persistent
- **WHEN** `Update` is called with any elapsed value
- **THEN** the internal rotation angle advances
- **THEN** `Update` returns `false` (never completes)

#### Scenario: Windfury animation is frame-rate independent
- **WHEN** `Update` is called with elapsed seconds
- **THEN** the rotation angle increments by `elapsed * 2.4`

### Requirement: KindWindfury effect kind
`effect.Kind` enum SHALL include `KindWindfury` for identifying windfury effects via `effect.List.Has`.

#### Scenario: Kind returns KindWindfury
- **WHEN** `Kind()` is called on a `Windfury` effect
- **THEN** it returns `KindWindfury`

### Requirement: Card renderer delegates windfury drawing to effect
`CardRenderer` SHALL use an `effect.Windfury` instance to render the windfury keyword visual. The `drawWindfury` method SHALL be removed from `CardRenderer`.

#### Scenario: Shop card with windfury
- **WHEN** `DrawShopCard` renders a minion with the Windfury keyword
- **THEN** the windfury effect's `DrawBehind` is called before the ellipse body
- **THEN** the windfury effect's `DrawFront` is called after badges

#### Scenario: Board card with windfury
- **WHEN** `DrawMinion` renders a minion with the Windfury keyword
- **THEN** the windfury effect's `DrawBehind` is called before the ellipse body
- **THEN** the windfury effect's `DrawFront` is called after badges

### Requirement: Windfury visual output is unchanged
The rendered pixels of the windfury effect SHALL be visually identical to the previous inline implementation. Orbit geometry, ribbon width, taper, wave modulation, colors, and alpha values MUST be preserved.

#### Scenario: Same visual at 60fps
- **WHEN** a windfury minion is rendered at 60fps
- **THEN** the wind ribbons match the previous output (orbit radius 1.05×ry, 39° inclination, 2 streaks per orbit, grey-white color RGBA{175,175,178})
