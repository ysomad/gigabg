# UI Architecture

## Overview

Minimal text-based UI for GIGA Battlegrounds. No interactive components - just text rendering for displaying game state.

## File Structure

```
ui/
├── DESIGN.md      # This document
├── scene.go       # Scene interface
├── menu.go        # Menu scene (lobby ID input)
├── game.go        # Game scene (main gameplay)
├── draw.go        # Text drawing helpers
└── layout.go      # Screen regions and constants
```

## Scenes

### Scene Interface

```go
type Scene interface {
    Update() error
    Draw(screen *ebiten.Image)
}
```

### Menu Scene

Simple lobby connection screen:
- Text input for lobby ID
- Join button
- Transitions to Game scene on successful connection

### Game Scene

Main gameplay display. Renders all game state as text.

## Screen Layout

Resolution: 1280x720

```
┌────────────────────────────────────────────────────────────┐
│ Turn 3 | Recruit Phase                              0:45   │  <- Header (turn, phase, timer)
│                                                            │
│ HP: 38    Gold: 5/7    Tier: 2    Upgrade: 5g              │  <- Player Stats
│                                                            │
│ Opponents: [P2: 40 HP] [P3: 35 HP] [P4: 32 HP] ...         │  <- Opponent Bar
│────────────────────────────────────────────────────────────│
│ SHOP (Tier 2)                                              │  <- Shop Header
│   1. Imp (1/1) Demon                                       │
│   2. Micro Machine (1/2) Mech                              │  <- Shop Cards
│   3. Sellemental (2/2) Elemental                           │
│────────────────────────────────────────────────────────────│
│ BOARD (2/7)                                                │  <- Board Header
│   1. Rockpool Hunter (2/3) Murloc                          │
│   2. Alleycat (1/1) Beast                                  │  <- Board Minions
└────────────────────────────────────────────────────────────┘
```

## Layout Regions

| Region | Y Start | Content |
|--------|---------|---------|
| Header | 20 | Turn number, phase, timer |
| Stats | 60 | HP, Gold, Tier, Upgrade cost |
| Opponents | 100 | Other players with HP |
| Shop | 160 | Shop tier header + card list |
| Board | 360 | Board count header + minion list |

## Data Requirements

### From GameClient

```go
type GameClient interface {
    Player() *game.Player    // HP, Gold, MaxGold, ShopTier, Board, Shop
    Turn() uint32
    Phase() game.Phase
    Timer() time.Duration    // TODO: add to client
    Opponents() []Opponent   // TODO: add to client
}
```

### Player Stats
- `HP` - Current health points
- `Gold` - Current gold
- `MaxGold` - Maximum gold this turn
- `ShopTier` - Current shop tier (1-6)
- `UpgradeCost()` - Cost to upgrade shop

### Opponent Info
- `ID` - Player identifier
- `HP` - Current health points

### Shop Cards
- `Name` - Card name
- `Attack` / `Health` - Stats
- `Tribe` - Card tribe

### Board Minions
- Same as shop cards
- Position index (1-7)

## Text Formatting

### Header Line
```
Turn {turn} | {phase}                              {mm:ss}
```

### Stats Line
```
HP: {hp}    Gold: {gold}/{maxGold}    Tier: {tier}    Upgrade: {cost}g
```

### Opponents Line
```
Opponents: [P1: {hp} HP] [P2: {hp} HP] ...
```

### Shop Card Line
```
  {index}. {name} ({attack}/{health}) {tribe}
```

### Board Minion Line
```
  {index}. {name} ({attack}/{health}) {tribe}
```

## Colors

| Element | Color (RGBA) |
|---------|--------------|
| Background | 20, 20, 30, 255 |
| Header text | 200, 200, 200, 255 |
| HP | 255, 100, 100, 255 |
| Gold | 255, 215, 0, 255 |
| Tier | 100, 200, 255, 255 |
| Timer | 255, 255, 255, 255 |
| Section header | 150, 150, 150, 255 |
| Card text | 255, 255, 255, 255 |
| Tribe text | 128, 128, 128, 255 |

## Future Additions

- Hand display (below board)
- Combat phase view
- Card interactions (click to buy/sell)
- Drag to reposition minions
- Animations
