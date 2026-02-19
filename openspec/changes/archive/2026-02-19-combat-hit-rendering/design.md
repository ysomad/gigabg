## Context

The current `HitDamage` effect (`ui/effect/hitdamage.go`) renders a gold circle with a damage number ("-X") on the target minion. Combat already emits two `DamageEvent`s per collision (one per direction via `dealDamage`), so both cards already receive a hit indicator. The problem is visual: there is no distinction between a normal damage hit and a poison hit — both show the same gold circle with a number.

The `DamageEvent` carries `Source` and `Target` combat IDs. The client's `combatBoard` holds both boards, so it can look up the source minion and check its keywords when processing a `DamageEvent`.

## Goals / Non-Goals

**Goals:**
- Refactor `HitDamage` into a `HitIndicator` type with a `HitType` discriminator
- Two visual variants: damage (gold/yellow + number) and poison (green + skull)
- Automatically select variant based on source minion's Poisonous/Venomous keyword

**Non-Goals:**
- Changing the wire format (`DamageEvent` stays the same)
- Modifying combat logic in `game/combat.go`
- Adding new combat event types

## Decisions

### 1. Resolve poison from client-side board state, not wire format

**Decision:** `applyDamage` looks up the source minion via `ev.Source` and checks `HasKeyword(Poisonous)` or `HasKeyword(Venomous)` to select the hit type. No changes to `DamageEvent`.

**Why:** The event ordering guarantees correctness — `DamageEvent` is emitted inside `dealDamage`, while `RemoveKeywordEvent` for Venomous is emitted later in `applyPoison`. So when the client processes a `DamageEvent`, the source minion still has its poison keywords on the client board. This avoids a wire format change and keeps the domain events minimal.

**Alternative considered:** Adding a `Poison bool` field to `DamageEvent`. Rejected because it couples domain events to UI presentation, and the client already has the state needed.

### 2. Rename HitDamage → HitIndicator with HitType enum

**Decision:** Replace `HitDamage` with `HitIndicator`. Add a `HitType` enum (`HitTypeDamage`, `HitTypePoison`). The constructor takes `HitType` to select visual config (colors, content). Rename `KindHitDamage` → `KindHitIndicator`.

**Why:** Clean separation of hit variant logic from animation mechanics. Same pop-in/hold/fade-out animation, different visual config per type.

### 3. Poison indicator draws a skull icon, not a number

**Decision:** `HitTypePoison` renders a green circle with a skull icon (drawn procedurally with vector shapes — circle head, eye sockets, jaw). No damage number shown — poison kills regardless of amount.

**Why:** Visually distinct from damage hits. Skull communicates "instant kill" without needing text. Procedural drawing keeps it consistent with existing keyword visuals (vial, shield, etc.) and avoids loading image assets.

**Alternative considered:** Showing both number and skull. Rejected — the number is meaningless for poison kills (the minion dies regardless of health).

## Risks / Trade-offs

- **Skull drawing complexity** → Keep it simple: circle + two eye dots + simple jaw line. Can iterate on the look after initial implementation.
- **Venomous keyword removal timing** → Mitigated by event ordering (DamageEvent always precedes RemoveKeywordEvent for the same attack). If event ordering ever changes in the combat engine, this would break. Low risk since events are emitted sequentially in `resolveAttack`.
