### Requirement: Tooltip appears to the right of the hovered card
The hover tooltip SHALL be positioned immediately to the right of the hovered card's bounding rect. The tooltip's left edge SHALL be offset from the card's right edge by a small gap. The tooltip's vertical center SHALL align with the hovered card's vertical center.

#### Scenario: Card hovered on board
- **WHEN** a player hovers over a board card
- **THEN** the tooltip appears with its left edge to the right of the card's right edge
- **THEN** the tooltip is vertically centered relative to the hovered card

#### Scenario: Card hovered in shop
- **WHEN** a player hovers over a shop card
- **THEN** the tooltip appears with its left edge to the right of the card's right edge
- **THEN** the tooltip is vertically centered relative to the hovered card

#### Scenario: Tooltip would overflow screen right edge
- **WHEN** a hovered card is near the right edge of the screen
- **THEN** the tooltip SHALL appear to the left of the hovered card instead

#### Scenario: Tooltip would overflow screen vertically
- **WHEN** the vertically centered tooltip would extend beyond the top or bottom screen edge
- **THEN** the tooltip SHALL be clamped to remain within screen bounds

### Requirement: No full-screen dim overlay
The hover tooltip SHALL NOT render a full-screen semi-transparent overlay behind the tooltip card.

#### Scenario: Tooltip displayed without overlay
- **WHEN** a tooltip is visible
- **THEN** the game board, shop, and all other UI elements remain fully visible with no dimming

### Requirement: Hover delay before tooltip appears
The tooltip SHALL only appear after the cursor remains over the same card for a configurable delay duration. The delay SHALL be at least 300ms.

#### Scenario: Quick mouse pass does not trigger tooltip
- **WHEN** the cursor enters a card rect and leaves within the delay duration
- **THEN** no tooltip is displayed

#### Scenario: Sustained hover triggers tooltip
- **WHEN** the cursor enters a card rect and remains for the full delay duration
- **THEN** the tooltip becomes visible

### Requirement: Delay resets when hovered card changes
The hover delay timer SHALL reset to zero when the cursor moves from one card to a different card, or when the cursor leaves all cards entirely.

#### Scenario: Moving between cards resets timer
- **WHEN** the cursor moves from card A to card B before the delay elapses
- **THEN** the delay timer resets and starts counting from zero for card B

#### Scenario: Leaving all cards resets timer
- **WHEN** the cursor moves off all cards
- **THEN** the delay timer resets to zero and the tooltip is hidden immediately
