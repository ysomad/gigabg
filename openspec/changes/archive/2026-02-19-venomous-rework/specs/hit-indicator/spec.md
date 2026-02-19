## MODIFIED Requirements

### Requirement: Poison hit indicator shown only on poison death
The poison hit indicator SHALL only appear when a minion dies from Venomous. When `markDying` processes a `DeathEvent` with `DeathReasonPoison`, it SHALL add a `HitTypePoison` indicator to the dying minion before the poison drip effect. `DeathReasonPoison` is set exclusively by the Venomous keyword combat path.

#### Scenario: Minion dies from Venomous
- **WHEN** a `DeathEvent` is processed with `DeathReasonPoison`
- **THEN** the dying minion receives a `HitTypePoison` indicator
- **AND** the poison drip effect is added after the indicator

#### Scenario: Venomous minion hits divine shield
- **WHEN** a Venomous minion attacks a target with Divine Shield
- **THEN** the shield breaks (no `DamageEvent` or `DeathEvent` emitted)
- **AND** no poison hit indicator is shown

#### Scenario: Minion dies from normal damage
- **WHEN** a `DeathEvent` is processed with `DeathReasonDamage`
- **THEN** no poison hit indicator is added
- **AND** only the damage hit indicator from `applyDamage` is visible
