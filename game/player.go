package game

import (
	"math/rand/v2"
	"slices"

	"github.com/ysomad/gigabg/errors"
)

const (
	ErrNotEnoughGold   errors.Error = "not enough gold"
	ErrBoardFull       errors.Error = "board is full"
	ErrHandFull        errors.Error = "hand is full"
	ErrInvalidIndex    errors.Error = "invalid index"
	ErrMaxTier         errors.Error = "already at max tier"
	ErrNotASpell       errors.Error = "card is not a spell"
	ErrDiscoverPending errors.Error = "discover already pending"
	ErrNoDiscover      errors.Error = "no discover options"
)

type Player struct {
	id              string
	hp              int
	gold            int
	maxGold         int
	shop            Shop
	board           Board  // minions on board
	hand            []Card // can hold minions and spells
	discoverOptions []Card // pending discover choices
}

func NewPlayer(id string) *Player {
	return &Player{
		id:      id,
		hp:      40,
		gold:    InitialGold,
		maxGold: MaxGold,
		shop:    Shop{tier: 1},
		board:   NewBoard(MaxBoardSize),
		hand:    make([]Card, 0, MaxHandSize),
	}
}

func (p *Player) ID() string   { return p.id }
func (p *Player) HP() int      { return p.hp }
func (p *Player) Gold() int    { return p.gold }
func (p *Player) MaxGold() int { return p.maxGold }
func (p *Player) Shop() Shop   { return p.shop }

// Hand returns a copy of the player's hand.
func (p *Player) Hand() []Card { return slices.Clone(p.hand) }

// HandSize returns the number of cards in hand.
func (p *Player) HandSize() int { return len(p.hand) }

// DiscoverOptions returns a copy of the pending discover choices.
func (p *Player) DiscoverOptions() []Card { return slices.Clone(p.discoverOptions) }

// HasDiscover returns true if the player has pending discover options.
func (p *Player) HasDiscover() bool { return len(p.discoverOptions) > 0 }

// Board returns a copy of the player's board.
func (p *Player) Board() Board { return p.board.Clone() }

// BoardSize returns the number of minions on the board.
func (p *Player) BoardSize() int { return p.board.Len() }

// Alive returns true if the player has HP remaining.
func (p *Player) Alive() bool {
	return p.hp > 0
}

// StartTurn prepares the player for a new turn.
func (p *Player) StartTurn(pool *CardPool, turn int) {
	if turn > 1 && p.maxGold < MaxGold {
		p.maxGold++
	}
	p.gold = p.maxGold

	p.shop.StartTurn(pool)
}

// TakeDamage reduces player HP and returns true if player is dead.
func (p *Player) TakeDamage(damage int) bool {
	p.hp -= damage
	return p.hp <= 0
}

// BuyCard buys a card from the shop and adds it to hand.
func (p *Player) BuyCard(shopIndex int) error {
	if p.gold < MinionPrice {
		return ErrNotEnoughGold
	}
	if len(p.hand) >= MaxHandSize {
		return ErrHandFull
	}

	card, err := p.shop.BuyCard(shopIndex)
	if err != nil {
		return err
	}

	p.hand = append(p.hand, card)
	p.gold -= MinionPrice
	return nil
}

// SellMinion sells a minion from the board for gold and returns it to the pool.
func (p *Player) SellMinion(boardIndex int, pool *CardPool) error {
	if boardIndex < 0 || boardIndex >= p.board.Len() {
		return ErrInvalidIndex
	}

	minion := p.board.RemoveMinion(boardIndex)
	pool.ReturnCard(minion)

	p.gold += SellValue
	if p.gold > MaxGold {
		p.gold = MaxGold
	}
	return nil
}

// PlaceMinion moves a minion from hand to board.
// When a golden minion is placed, a Triple Reward spell is added to hand.
func (p *Player) PlaceMinion(handIndex, boardPosition int, cards CardStore) error {
	if handIndex < 0 || handIndex >= len(p.hand) {
		return ErrInvalidIndex
	}
	if p.board.IsFull() {
		return ErrBoardFull
	}

	card := p.hand[handIndex]
	minion, ok := card.(*Minion)
	if !ok {
		return ErrInvalidIndex
	}

	p.hand = append(p.hand[:handIndex], p.hand[handIndex+1:]...)
	p.board.PlaceMinion(minion, boardPosition)

	// Golden minion placement grants Triple Reward spell
	if minion.IsGolden() {
		if tmpl := cards.ByTemplateID("triple_reward"); tmpl != nil && len(p.hand) < MaxHandSize {
			p.hand = append(p.hand, NewSpell(tmpl))
		}
	}

	return nil
}

// RemoveMinion moves a minion from board to hand.
func (p *Player) RemoveMinion(boardIndex int) error {
	if boardIndex < 0 || boardIndex >= p.board.Len() {
		return ErrInvalidIndex
	}
	if len(p.hand) >= MaxHandSize {
		return ErrHandFull
	}

	minion := p.board.RemoveMinion(boardIndex)
	p.hand = append(p.hand, minion)
	return nil
}

// UpgradeShop upgrades the shop tier.
func (p *Player) UpgradeShop() error {
	if p.shop.Tier() >= Tier6 {
		return ErrMaxTier
	}

	cost := p.shop.UpgradeCost()
	if p.gold < cost {
		return ErrNotEnoughGold
	}

	p.gold -= cost
	p.shop.Upgrade()
	return nil
}

// RefreshShop refreshes the shop for gold, returning old cards to pool.
func (p *Player) RefreshShop(pool *CardPool) error {
	if p.gold < ShopRefreshCost {
		return ErrNotEnoughGold
	}

	p.gold -= ShopRefreshCost
	p.shop.Refresh(pool)
	return nil
}

// FreezeShop toggles the shop freeze state.
func (p *Player) FreezeShop() {
	p.shop.Freeze()
}

// CheckTriples scans hand + board for 3 non-golden copies of the same minion.
// If found, removes all 3, creates a golden minion (2x stats) in hand.
// Returns true if a triple was found and combined.
func (p *Player) CheckTriples() bool {
	type loc struct {
		board bool
		index int
	}

	groups := make(map[string][]loc)
	minions := make(map[string]*Minion)

	for i, m := range p.board.Minions() {
		if m.IsGolden() {
			continue
		}
		tid := m.TemplateID()
		groups[tid] = append(groups[tid], loc{board: true, index: i})
		minions[tid] = m
	}
	for i, c := range p.hand {
		m, ok := c.(*Minion)
		if !ok || m.IsGolden() {
			continue
		}
		tid := m.TemplateID()
		groups[tid] = append(groups[tid], loc{index: i})
		minions[tid] = m
	}

	for tid, locs := range groups {
		if len(locs) < 3 {
			continue
		}

		var boardIdxs, handIdxs []int
		for _, l := range locs[:3] {
			if l.board {
				boardIdxs = append(boardIdxs, l.index)
			} else {
				handIdxs = append(handIdxs, l.index)
			}
		}

		// Remove in descending order to preserve indices.
		slices.Sort(boardIdxs)
		slices.Reverse(boardIdxs)
		slices.Sort(handIdxs)
		slices.Reverse(handIdxs)

		for _, idx := range boardIdxs {
			p.board.RemoveMinion(idx)
		}
		for _, idx := range handIdxs {
			p.hand = slices.Delete(p.hand, idx, idx+1)
		}

		p.hand = append(p.hand, minions[tid].ToGolden())
		return true
	}

	return false
}

// PlaySpell plays a spell from hand.
func (p *Player) PlaySpell(handIndex int, pool *CardPool) error {
	if handIndex < 0 || handIndex >= len(p.hand) {
		return ErrInvalidIndex
	}

	spell, ok := p.hand[handIndex].(*Spell)
	if !ok {
		return ErrNotASpell
	}

	if p.discoverOptions != nil {
		return ErrDiscoverPending
	}

	// Remove spell from hand
	p.hand = append(p.hand[:handIndex], p.hand[handIndex+1:]...)

	// Execute spell effect
	effect := spell.Template().SpellEffect
	if effect != nil && effect.Type == EffectDiscoverCard {
		discoverTier := min(p.shop.Tier()+1, Tier6)
		p.discoverOptions = pool.RollExactTier(discoverTier, nil)
	}

	return nil
}

// DiscoverPick picks one of the discover options and adds it to hand.
// Unpicked options are returned to the pool.
func (p *Player) DiscoverPick(index int, pool *CardPool) error {
	if p.discoverOptions == nil {
		return ErrNoDiscover
	}
	if index < 0 || index >= len(p.discoverOptions) {
		return ErrInvalidIndex
	}
	if len(p.hand) >= MaxHandSize {
		return ErrHandFull
	}

	// Return unpicked options to pool
	for i, c := range p.discoverOptions {
		if i != index {
			pool.ReturnCard(c)
		}
	}

	p.hand = append(p.hand, p.discoverOptions[index])
	p.discoverOptions = nil
	return nil
}

// ResolveDiscover auto-picks a random discover option for the player.
// If hand is full, all options are returned to the pool.
// Always clears discover options when done.
func (p *Player) ResolveDiscover(pool *CardPool) {
	defer func() { p.discoverOptions = nil }()

	if len(p.discoverOptions) == 0 {
		return
	}

	if len(p.hand) >= MaxHandSize {
		pool.ReturnCards(p.discoverOptions)
		return
	}

	idx := rand.IntN(len(p.discoverOptions))
	p.hand = append(p.hand, p.discoverOptions[idx])

	for i, c := range p.discoverOptions {
		if i != idx {
			pool.ReturnCard(c)
		}
	}
}

// ReorderBoard reorders the board based on the given indices.
func (p *Player) ReorderBoard(order []int) error {
	return p.board.Reorder(order)
}
