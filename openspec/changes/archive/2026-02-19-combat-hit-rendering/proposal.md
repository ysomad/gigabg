## Why

When two minions collide in combat, the hit indicator (damage number) only appears on the target. Both minions deal damage to each other simultaneously, but only one side gets visual feedback. Additionally, there is no visual distinction between a normal damage hit and a poison kill — the same gold circle is shown regardless.

## What Changes

- Extract the current `HitDamage` effect into a dedicated `HitIndicator` type with a `HitType` discriminator
- Add two visual variants:
  - **Damage** — yellow/gold background circle with the numeric damage amount (current behavior, refined)
  - **Poison** — green background circle with a skull icon, shown when the attacker has Poisonous or Venomous
- Show hit indicators on **both** the attacker and the defender after a combat collision (each card shows the damage it received)
- Selection logic: if the minion dealing damage has `Poisonous` or `Venomous` keyword, the target shows a poison hit indicator; otherwise it shows a damage hit indicator

## Capabilities

### New Capabilities

- `hit-indicator`: Defines the hit indicator effect type, its two visual variants (damage, poison), rendering rules, selection logic based on attacker keywords, and dual-card display on combat collision

### Modified Capabilities

_(none — existing specs are unaffected)_

## Impact

- `ui/effect/` — new or refactored effect type replacing `HitDamage`
- `ui/scene/combatboard.go` — `applyDamage` must add hit indicators to both attacker and defender, passing attacker keyword info to select variant
- `game/combatevent.go` — `DamageEvent` may need to carry source keyword context (Poisonous/Venomous) so the client can choose the correct indicator, or the client resolves this from board state
