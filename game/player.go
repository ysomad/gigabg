package game

import "github.com/ysomad/gigabg/errorsx"

const (
	MaxPlayers   = 2
	MaxBoardSize = 7
	MaxHandSize  = 10

	InitialGold = 3
	MaxGold     = 10

	BuyCost     = 3
	SellValue   = 1
	RefreshCost = 1
)

const (
	ErrNotEnoughGold   errorsx.Error = "not enough gold"
	ErrBoardFull       errorsx.Error = "board is full"
	ErrHandFull        errorsx.Error = "hand is full"
	ErrInvalidIndex    errorsx.Error = "invalid index"
	ErrMaxTier         errorsx.Error = "already at max tier"
	ErrNotASpell       errorsx.Error = "card is not a spell"
	ErrDiscoverPending errorsx.Error = "discover already pending"
	ErrNoDiscover      errorsx.Error = "no discover options"
	ErrCannotSellSpell errorsx.Error = "cannot sell spells"
)

type Player struct {
	ID      string
	HP      int
	Gold    int
	MaxGold int

	Shop Shop

	Board           []*Minion // only minions on board
	Hand            []Card    // can hold minions and spells
	DiscoverOptions []Card    // pending discover choices
}

func NewPlayer(id string) *Player {
	return &Player{
		ID:      id,
		HP:      40,
		Gold:    InitialGold,
		MaxGold: MaxGold,
		Shop:    Shop{tier: 1},
		Board:   make([]*Minion, 0, MaxBoardSize),
		Hand:    make([]Card, 0, MaxHandSize),
	}
}

// StartTurn prepares the player for a new turn.
func (p *Player) StartTurn(pool *CardPool, turn int) {
	if turn > 1 && p.MaxGold < MaxGold {
		p.MaxGold++
	}
	p.Gold = p.MaxGold

	p.Shop.StartTurn(pool)
}

// TakeDamage reduces player HP and returns true if player is dead.
func (p *Player) TakeDamage(damage int) bool {
	p.HP -= damage
	return p.HP <= 0
}

// BuyCard buys a card from the shop and adds it to hand.
func (p *Player) BuyCard(shopIndex int) error {
	if p.Gold < BuyCost {
		return ErrNotEnoughGold
	}
	if len(p.Hand) >= MaxHandSize {
		return ErrHandFull
	}

	card, err := p.Shop.BuyCard(shopIndex)
	if err != nil {
		return err
	}

	p.Hand = append(p.Hand, card)
	p.Gold -= BuyCost
	return nil
}

// SellCard sells a card from hand for gold and returns it to the pool.
func (p *Player) SellCard(handIndex int, pool *CardPool) error {
	if handIndex < 0 || handIndex >= len(p.Hand) {
		return ErrInvalidIndex
	}
	if p.Hand[handIndex].IsSpell() {
		return ErrCannotSellSpell
	}

	card := p.Hand[handIndex]
	p.Hand = append(p.Hand[:handIndex], p.Hand[handIndex+1:]...)
	pool.ReturnCard(card)

	p.Gold += SellValue
	if p.Gold > MaxGold {
		p.Gold = MaxGold
	}
	return nil
}

// PlaceMinion moves a minion from hand to board.
// When a golden minion is placed, a Triple Reward spell is added to hand.
func (p *Player) PlaceMinion(handIndex, boardPosition int, cards CardStore) error {
	if handIndex < 0 || handIndex >= len(p.Hand) {
		return ErrInvalidIndex
	}
	if len(p.Board) >= MaxBoardSize {
		return ErrBoardFull
	}

	card := p.Hand[handIndex]
	minion, ok := card.(*Minion)
	if !ok {
		return ErrInvalidIndex
	}

	p.Hand = append(p.Hand[:handIndex], p.Hand[handIndex+1:]...)

	if boardPosition < 0 {
		boardPosition = 0
	}
	if boardPosition > len(p.Board) {
		boardPosition = len(p.Board)
	}

	p.Board = append(p.Board[:boardPosition], append([]*Minion{minion}, p.Board[boardPosition:]...)...)

	// Golden minion placement grants Triple Reward spell
	if minion.Golden() {
		if tmpl := cards.ByTemplateID("triple_reward"); tmpl != nil && len(p.Hand) < MaxHandSize {
			p.Hand = append(p.Hand, NewSpell(tmpl))
		}
	}

	return nil
}

// RemoveMinion moves a minion from board to hand.
func (p *Player) RemoveMinion(boardIndex int) error {
	if boardIndex < 0 || boardIndex >= len(p.Board) {
		return ErrInvalidIndex
	}
	if len(p.Hand) >= MaxHandSize {
		return ErrHandFull
	}

	minion := p.Board[boardIndex]
	p.Board = append(p.Board[:boardIndex], p.Board[boardIndex+1:]...)
	p.Hand = append(p.Hand, minion)
	return nil
}

// UpgradeShop upgrades the shop tier.
func (p *Player) UpgradeShop() error {
	if p.Shop.Tier() >= Tier6 {
		return ErrMaxTier
	}

	cost := p.Shop.UpgradeCost()
	if p.Gold < cost {
		return ErrNotEnoughGold
	}

	p.Gold -= cost
	p.Shop.Upgrade()
	return nil
}

// RefreshShop refreshes the shop for gold, returning old cards to pool.
func (p *Player) RefreshShop(pool *CardPool) error {
	if p.Gold < RefreshCost {
		return ErrNotEnoughGold
	}

	p.Gold -= RefreshCost
	p.Shop.Refresh(pool)
	return nil
}

// FreezeShop toggles the shop freeze state.
func (p *Player) FreezeShop() {
	p.Shop.Freeze()
}

// CheckTriples scans hand + board for 3 non-golden copies of the same minion.
// If found, removes all 3, creates a golden minion (2x stats) in hand,
// and adds a Triple Reward spell to hand if there's room.
// Returns true if a triple was found and combined.
func (p *Player) CheckTriples() bool {
	type loc struct {
		fromBoard bool
		index     int
	}

	counts := make(map[string][]loc)

	for i, m := range p.Board {
		if m.Golden() {
			continue
		}
		counts[m.TemplateID()] = append(counts[m.TemplateID()], loc{fromBoard: true, index: i})
	}
	for i, c := range p.Hand {
		m, ok := c.(*Minion)
		if !ok || m.Golden() {
			continue
		}
		counts[m.TemplateID()] = append(counts[m.TemplateID()], loc{fromBoard: false, index: i})
	}

	for _, locs := range counts {
		if len(locs) < 3 {
			continue
		}

		// Save template before removal
		first := locs[0]
		var tmpl *CardTemplate
		if first.fromBoard {
			tmpl = p.Board[first.index].Template()
		} else {
			tmpl = p.Hand[first.index].(*Minion).Template()
		}

		// Remove in reverse index order to avoid shifting issues.
		// Separate board and hand removals.
		var boardIdxs, handIdxs []int
		for _, l := range locs[:3] {
			if l.fromBoard {
				boardIdxs = append(boardIdxs, l.index)
			} else {
				handIdxs = append(handIdxs, l.index)
			}
		}

		// Sort descending and remove
		sortDesc(boardIdxs)
		for _, idx := range boardIdxs {
			p.Board = append(p.Board[:idx], p.Board[idx+1:]...)
		}

		sortDesc(handIdxs)
		for _, idx := range handIdxs {
			p.Hand = append(p.Hand[:idx], p.Hand[idx+1:]...)
		}
		golden := NewMinion(tmpl)
		golden.SetGolden(true)
		golden.SetAttack(tmpl.Golden.Attack)
		golden.SetHealth(tmpl.Golden.Health)
		p.Hand = append(p.Hand, golden)

		return true
	}

	return false
}

// PlaySpell plays a spell from hand.
func (p *Player) PlaySpell(handIndex int, pool *CardPool) error {
	if handIndex < 0 || handIndex >= len(p.Hand) {
		return ErrInvalidIndex
	}

	spell, ok := p.Hand[handIndex].(*Spell)
	if !ok {
		return ErrNotASpell
	}

	if p.DiscoverOptions != nil {
		return ErrDiscoverPending
	}

	// Remove spell from hand
	p.Hand = append(p.Hand[:handIndex], p.Hand[handIndex+1:]...)

	// Execute spell effect
	effect := spell.Template().SpellEffect
	if effect != nil && effect.Type == EffectDiscoverCard {
		discoverTier := min(p.Shop.Tier()+1, Tier6)
		p.DiscoverOptions = pool.RollExactTier(discoverTier, nil)
	}

	return nil
}

// DiscoverPick picks one of the discover options and adds it to hand.
// Unpicked options are returned to the pool.
func (p *Player) DiscoverPick(index int, pool *CardPool) error {
	if p.DiscoverOptions == nil {
		return ErrNoDiscover
	}
	if index < 0 || index >= len(p.DiscoverOptions) {
		return ErrInvalidIndex
	}
	if len(p.Hand) >= MaxHandSize {
		return ErrHandFull
	}

	// Return unpicked options to pool
	for i, c := range p.DiscoverOptions {
		if i != index {
			pool.ReturnCard(c)
		}
	}

	p.Hand = append(p.Hand, p.DiscoverOptions[index])
	p.DiscoverOptions = nil
	return nil
}

func sortDesc(s []int) {
	for i := range s {
		for j := i + 1; j < len(s); j++ {
			if s[j] > s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// ReorderBoard reorders the board based on the given indices.
func (p *Player) ReorderBoard(order []int) error {
	if len(order) != len(p.Board) {
		return ErrInvalidIndex
	}

	newBoard := make([]*Minion, len(p.Board))
	used := make(map[int]struct{}, len(p.Board))

	for i, idx := range order {
		if idx < 0 || idx >= len(p.Board) {
			return ErrInvalidIndex
		}
		if _, ok := used[idx]; ok {
			return ErrInvalidIndex
		}
		newBoard[i] = p.Board[idx]
		used[idx] = struct{}{}
	}

	p.Board = newBoard
	return nil
}
