## ADDED Requirements

### Requirement: HitType enum discriminates hit indicator variants
The system SHALL define a `HitType` enum with two values: `HitTypeDamage` (iota+1) and `HitTypePoison`. `HitType` SHALL be used by `HitIndicator` to select its visual presentation.

#### Scenario: HitType values start at iota+1
- **WHEN** `HitType` is defined
- **THEN** `HitTypeDamage` equals 1 and `HitTypePoison` equals 2

### Requirement: HitIndicator replaces HitDamage as the hit effect
The system SHALL replace the `HitDamage` effect with `HitIndicator`. The constructor SHALL accept `HitType`, `duration`, `damage` (int), and `boldFont`. `KindHitDamage` SHALL be renamed to `KindHitIndicator`. `HitIndicator` SHALL implement the `Effect` interface.

#### Scenario: Creating a damage hit indicator
- **WHEN** `NewHitIndicator(HitTypeDamage, duration, 3, font)` is called
- **THEN** a `HitIndicator` is returned with `Kind() == KindHitIndicator`
- **AND** it renders as the damage variant (gold circle with number)

#### Scenario: Creating a poison hit indicator
- **WHEN** `NewHitIndicator(HitTypePoison, duration, 0, font)` is called
- **THEN** a `HitIndicator` is returned with `Kind() == KindHitIndicator`
- **AND** it renders as the poison variant (green circle with skull)

### Requirement: Damage variant renders a gold circle with damage number
When `HitType` is `HitTypeDamage`, `DrawFront` SHALL render a gold/yellow background circle at the card center with a white "-X" damage number (where X is the damage amount). The indicator SHALL pop in with overshoot, hold, then fade out.

#### Scenario: Damage indicator visual
- **WHEN** a damage hit indicator is drawn
- **THEN** a gold circle (approx. RGBA 180,140,10) is rendered at card center
- **AND** white text "-X" (X = damage amount) is drawn centered on the circle
- **AND** the circle has a dark border stroke

#### Scenario: Damage indicator with zero or negative damage
- **WHEN** damage amount is <= 0
- **THEN** nothing is drawn

### Requirement: Poison variant renders a green circle with skull icon
When `HitType` is `HitTypePoison`, `DrawFront` SHALL render a green background circle at the card center with a white skull icon drawn procedurally. The same pop-in/hold/fade-out animation timing SHALL apply.

#### Scenario: Poison indicator visual
- **WHEN** a poison hit indicator is drawn
- **THEN** a green circle (approx. RGBA 30,160,50) is rendered at card center
- **AND** a white skull icon is drawn centered on the circle
- **AND** the circle has a dark border stroke

### Requirement: Hit indicator animation has pop-in and fade-out phases
Both variants SHALL share the same animation: pop-in with overshoot scale in the first 20% of lifetime, hold at normal scale, then fade out alpha in the last 30% of lifetime.

#### Scenario: Pop-in phase
- **WHEN** progress is in the first 20% of lifetime
- **THEN** the circle scale overshoots by ~30% and decreases toward 1.0

#### Scenario: Fade-out phase
- **WHEN** remaining time is in the last 30% of lifetime
- **THEN** alpha decreases from 255 toward 0

#### Scenario: Hold phase
- **WHEN** progress is between pop-in end and fade-out start
- **THEN** the indicator is rendered at full scale and full opacity

### Requirement: applyDamage selects hit type based on source minion keywords
When processing a `DamageEvent`, the combat board SHALL look up the source minion via `ev.Source`. If the source minion has `KeywordPoisonous` or `KeywordVenomous`, the target SHALL receive a poison hit indicator. Otherwise, the target SHALL receive a damage hit indicator.

#### Scenario: Attacker has Poisonous keyword
- **WHEN** a `DamageEvent` is processed
- **AND** the source minion has `KeywordPoisonous`
- **THEN** the target minion receives a `HitTypePoison` indicator

#### Scenario: Attacker has Venomous keyword
- **WHEN** a `DamageEvent` is processed
- **AND** the source minion has `KeywordVenomous`
- **THEN** the target minion receives a `HitTypePoison` indicator

#### Scenario: Attacker has no poison keyword
- **WHEN** a `DamageEvent` is processed
- **AND** the source minion has neither `KeywordPoisonous` nor `KeywordVenomous`
- **THEN** the target minion receives a `HitTypeDamage` indicator with the damage amount

#### Scenario: Source minion not found
- **WHEN** a `DamageEvent` is processed
- **AND** the source minion cannot be found on either board (e.g., already dead)
- **THEN** the target minion receives a `HitTypeDamage` indicator as fallback
