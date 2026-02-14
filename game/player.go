package game

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"

	"github.com/ysomad/gigabg/pkg/errors"
)

type PlayerID int32

func ParsePlayerID(s string) (PlayerID, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return PlayerID(n), nil
}

const (
	ErrNotEnoughGold        errors.Error = "not enough gold"
	ErrBoardFull            errors.Error = "board is full"
	ErrHandFull             errors.Error = "hand is full"
	ErrInvalidHandIndex     errors.Error = "invalid hand index"
	ErrInvalidBoardIndex    errors.Error = "invalid board index"
	ErrInvalidShopIndex     errors.Error = "invalid shop index"
	ErrInvalidDiscoverIndex errors.Error = "invalid discover index"
	ErrInvalidReorder       errors.Error = "invalid reorder"
	ErrMaxTier              errors.Error = "already at max tier"
	ErrNotAMinion           errors.Error = "card is not a minion"
	ErrNotASpell            errors.Error = "card is not a spell"
	ErrDiscoverPending      errors.Error = "discover already pending"
	ErrNoDiscover           errors.Error = "no discover options"
)

type Player struct {
	id        PlayerID
	hp        int
	gold      int
	maxGold   int
	placement int
	shop      Shop
	board     Board  // minions on board
	hand      Hand   // can hold minions and spells
	discovers []Card // pending discover options
}

func NewPlayer(id PlayerID) *Player {
	return &Player{
		id:      id,
		hp:      initialHP,
		gold:    initialGold,
		maxGold: maxGold,
		shop:    Shop{tier: Tier1, refreshCost: ShopRefreshCost},
		board:   NewBoard(maxBoardSize),
		hand:    NewHand(),
	}
}

func (p *Player) ID() PlayerID       { return p.id }
func (p *Player) HP() int            { return p.hp }
func (p *Player) Gold() int          { return p.gold }
func (p *Player) MaxGold() int       { return p.maxGold }
func (p *Player) Placement() int     { return p.placement }
func (p *Player) SetPlacement(n int) { p.placement = n }
func (p *Player) Shop() Shop         { return p.shop }

// Hand returns a copy of the player's hand cards.
func (p *Player) Hand() []Card { return p.hand.Cards() }

// HandSize returns the number of cards in hand.
func (p *Player) HandSize() int { return p.hand.Len() }

// Discovers returns a copy of the pending discover choices.
func (p *Player) Discovers() []Card { return slices.Clone(p.discovers) }

// HasDiscovers returns true if the player has pending discover options.
func (p *Player) HasDiscovers() bool { return len(p.discovers) > 0 }

// Board returns a copy of the player's board.
func (p *Player) Board() Board { return p.board.Clone() }

// BoardSize returns the number of minions on the board.
func (p *Player) BoardSize() int { return p.board.Len() }

// IsAlive returns true if the player has HP remaining.
func (p *Player) IsAlive() bool { return p.hp > 0 }

// StartTurn prepares the player for a new turn.
func (p *Player) StartTurn(pool *CardPool, turn int) {
	if turn > 1 && p.maxGold < maxGold {
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
func (p *Player) BuyCard(shopIdx int) error {
	if !p.shop.HasCardAt(shopIdx) {
		return ErrInvalidShopIndex
	}
	if p.hand.IsFull() {
		return ErrHandFull
	}

	cost := p.shop.cards[shopIdx].Template().Cost()
	if p.gold < cost {
		return ErrNotEnoughGold
	}

	card, err := p.shop.BuyCard(shopIdx)
	if err != nil {
		return fmt.Errorf("shop: %w", err)
	}

	p.hand.Add(card)
	p.gold -= cost

	return nil
}

// SellMinion sells a minion from the board for gold and returns it to the pool.
func (p *Player) SellMinion(boardIndex int, pool *CardPool) error {
	if !p.board.HasMinionAt(boardIndex) {
		return ErrInvalidBoardIndex
	}

	minion := p.board.RemoveMinion(boardIndex)
	pool.ReturnCard(minion)

	p.gold += minionSellValue
	if p.gold > maxGold {
		p.gold = maxGold
	}

	return nil
}

// CanPlayCard checks if the card at handIdx can be played and returns an error if not.
func (p *Player) CanPlayCard(handIdx int) error {
	if !p.hand.HasCardAt(handIdx) {
		return ErrInvalidHandIndex
	}
	card := p.hand.CardAt(handIdx)
	if card.IsMinion() && p.board.IsFull() {
		return ErrBoardFull
	}
	if card.IsSpell() && p.HasDiscovers() {
		return ErrDiscoverPending
	}
	return nil
}

func newEffectContext(p *Player, pool *CardPool) EffectContext {
	return EffectContext{
		Board:     &p.board,
		Hand:      &p.hand,
		Shop:      &p.shop,
		Pool:      pool,
		Discovers: &p.discovers,
	}
}

// PlayMinion moves a minion from hand to board and executes its Battlecry effects.
// If the minion has Magnetic and is placed to the left of a Mech, it fuses onto that Mech instead.
func (p *Player) PlayMinion(handIdx, boardIdx int, pool *CardPool) error {
	if !p.hand.HasCardAt(handIdx) {
		return ErrInvalidHandIndex
	}

	minion, ok := p.hand.CardAt(handIdx).(*Minion)
	if !ok {
		return ErrNotAMinion
	}

	target := p.board.MinionAt(boardIdx)
	magnetize := minion.CanMagnetizeTo(target)

	if !magnetize {
		if p.board.IsFull() {
			return ErrBoardFull
		}
		if !p.board.CanPlaceAt(boardIdx) {
			return ErrInvalidBoardIndex
		}
	}

	p.hand.RemoveCard(handIdx)

	if magnetize {
		mergeMinions(minion, target)
	} else {
		p.board.PlaceMinion(minion, boardIdx)
	}

	ctx := newEffectContext(p, pool).WithSource(minion)
	for e := range minion.EffectsByTrigger(TriggerBattlecry, TriggerGolden) {
		e.Apply(ctx)
	}

	return nil
}

// mergeMinions fuses source stats, keywords and effects onto target.
func mergeMinions(source, target *Minion) {
	target.attack += source.attack
	target.health += source.health
	target.keywords.Merge(source.keywords)
	target.effects = append(target.effects, source.effects...)
}

// RemoveMinion moves a minion from board to hand.
func (p *Player) RemoveMinion(boardIndex int) error {
	if !p.board.HasMinionAt(boardIndex) {
		return ErrInvalidBoardIndex
	}
	if p.hand.IsFull() {
		return ErrHandFull
	}

	minion := p.board.RemoveMinion(boardIndex)
	p.hand.Add(minion)
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
	if p.gold < p.shop.refreshCost {
		return ErrNotEnoughGold
	}

	p.gold -= p.shop.refreshCost
	p.shop.Refresh(pool)
	return nil
}

// FreezeShop toggles the shop freeze state.
func (p *Player) FreezeShop() { p.shop.Freeze() }

// CheckTriples scans hand + board for 3 non-golden copies of the same minion.
// If found, removes all 3, merges first 2 into a golden minion in hand.
// The 3rd copy is consumed without contributing stats.
// Returns true if a triple was found and combined.
func (p *Player) CheckTriples() bool {
	type loc struct {
		board bool
		index int
	}

	groups := make(map[string][]loc)
	instances := make(map[string][]*Minion)

	for i, m := range p.board.Minions() {
		if m.IsGolden() {
			continue
		}
		tid := m.TemplateID()
		groups[tid] = append(groups[tid], loc{board: true, index: i})
		instances[tid] = append(instances[tid], m)
	}
	for i := range p.hand.Len() {
		m, ok := p.hand.CardAt(i).(*Minion)
		if !ok || m.IsGolden() {
			continue
		}
		tid := m.TemplateID()
		groups[tid] = append(groups[tid], loc{index: i})
		instances[tid] = append(instances[tid], m)
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
		p.hand.RemoveDesc(handIdxs)

		p.hand.Add(instances[tid][0].MergeGolden(instances[tid][1]))
		return true
	}

	return false
}

// PlaySpell plays a spell from hand.
func (p *Player) PlaySpell(handIdx int, pool *CardPool) error {
	if err := p.CanPlayCard(handIdx); err != nil {
		return err
	}

	spell, ok := p.hand.CardAt(handIdx).(*Spell)
	if !ok {
		return ErrNotASpell
	}

	p.hand.RemoveCard(handIdx)

	for e := range EffectsByTrigger(spell.Template().Effects(), TriggerSpell) {
		e.Apply(newEffectContext(p, pool))
	}

	return nil
}

// DiscoverPick picks one of the discover options and adds it to hand.
// Unpicked options are returned to the pool.
func (p *Player) DiscoverPick(index int, pool *CardPool) error {
	if p.discovers == nil {
		return ErrNoDiscover
	}
	if index < 0 || index >= len(p.discovers) {
		return ErrInvalidDiscoverIndex
	}
	if p.hand.IsFull() {
		return ErrHandFull
	}

	// Return unpicked options to pool
	for i, c := range p.discovers {
		if i != index {
			pool.ReturnCard(c)
		}
	}

	p.hand.Add(p.discovers[index])
	p.discovers = nil
	return nil
}

// ResolveDiscover auto-picks a random discover option for the player.
// If hand is full, all options are returned to the pool.
// Always clears discover options when done.
func (p *Player) ResolveDiscover(pool *CardPool) {
	defer func() { p.discovers = nil }()

	if len(p.discovers) == 0 {
		return
	}

	if p.hand.IsFull() {
		pool.ReturnCards(p.discovers)
		return
	}

	idx := rand.IntN(len(p.discovers)) //nolint:gosec // game logic, not crypto
	p.hand.Add(p.discovers[idx])

	for i, c := range p.discovers {
		if i != idx {
			pool.ReturnCard(c)
		}
	}
}

// ReorderBoard reorders the board based on the given indices.
func (p *Player) ReorderBoard(order []int) error { return p.board.Reorder(order) }

// ReorderShop reorders the shop cards based on the given indices.
func (p *Player) ReorderShop(order []int) error { return p.shop.Reorder(order) }
