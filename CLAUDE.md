# GIGA Battlegrounds

Heartstone Battlegrounds like game written in Go using Ebiten engine.

## Entities and mechanics

### Game
- 8 players per lobby
- Turns alternate between Recruit Phase and Combat Phase
- Starts with recruit phase
- Last player standing wins

### Player
- Starts with 40 HP
- Maximum 10 gold per turn (starts at 3, +1 each turn)
- Maximum 7 minions on board
- Maximum 10 cards in hand

### Shop
- Shop where players buy/sell cards
- Refresh: 1 gold for new card selection
- Shop Tiers 1-6, higher tiers have stronger cards
- Upgrade cost decreases each turn

### Minion
- Has Attack (AP) and Health (HP)
- May have one or multiple Keywords)
- Belongs to a Tribe (or no tribe)
- Can be Tier 1-6
- Golden: combine 3 copies for 2x stats + bonus effect

### Tribes
Beast, Demon, Dragon, Elemental, Mech, Murloc, Naga, Pirate, Quilboar, Undead (+ Neutral)

### Keywords
Static keywords (affect combat behavior):
- **Taunt**: must be attacked first
- **Divine Shield**: ignores first damage instance
- **Windfury**: attacks twice per combat
- **Reborn**: returns with 1 HP after first death
- **Poisonous**: kills any minion it damages
- **Cleave**: deals damage to adjacent minions when attacking
- **Stealth**: cannot be targeted until it attacks
- **Immune**: cannot take damage
- **Magnetic**: can merge with adjacent Mech

Triggered abilities (have Effect payloads):
- **Battlecry**: triggers when played from hand
- **Deathrattle**: triggers on death
- **Avenge**: triggers after friendly minions die (threshold-based)
- **Start of Combat**: triggers at combat start
- **End of Turn**: triggers at end of recruit phase

### Effects
Effects define what triggered abilities do. Effect types:
- **BuffStats**: give +Attack/+Health
- **GiveKeyword**: give Divine Shield, Taunt, etc.
- **Summon**: spawn minions on board
- **DealDamage**: deal damage to targets
- **Destroy**: destroy target minion
- **AddCard**: add card to hand (random, specific, or filtered by tribe/tier)

Target types: Self, AllFriendly, AllEnemy, RandomFriendly, RandomEnemy, LeftFriendly, RightFriendly, LeftmostFriendly, RightmostFriendly

Targets can be filtered by Tribe, Tier, or Keyword. Can also exclude self.

### Stat Persistence
- **Base stats**: permanent stats visible in Recruit phase
- **Combat state**: damage taken during combat resets after combat ends
- Effects with `Persistent: true` modify base stats (permanent)
- Effects with `Persistent: false` only affect combat (temporary)

### Recruit phase
- 60s for each player, with option to specify dynamically

### Combat
- Automated battle between two players' boards
- Minions attack left-to-right, target random enemy (Taunt priority)
- Loser takes damage = opponent's shop tier + sum of surviving minions' tiers
- **Combat damage is temporary**: surviving minions restore to pre-combat stats (health resets)
- After combat players goes to Recruit phase and their board is restored

## Styleguide
See `./STYLE.md`

## Ebiten
- Use `vector.FillRect` instead of `vector.DrawFilledRect`
- Use `vector.FillCircle` instead of `vector.DrawFilledCircle`
