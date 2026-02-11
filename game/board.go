package game

import (
	"math/rand/v2"
	"slices"
)

// Board holds minions on a player's board.
type Board struct {
	minions []*Minion
}

// NewBoard creates a board with the given initial capacity.
func NewBoard(size int) Board {
	return Board{minions: make([]*Minion, 0, size)}
}

// Len returns the number of minions on the board.
func (b Board) Len() int { return len(b.minions) }

// IsFull returns true if the board has reached MaxBoardSize.
func (b Board) IsFull() bool { return len(b.minions) >= MaxBoardSize }

// GetMinion returns the minion at the given index, or nil if out of range.
func (b Board) GetMinion(i int) *Minion {
	if i < 0 || i >= len(b.minions) {
		return nil
	}
	return b.minions[i]
}

// Minions returns a copy of the minions slice.
func (b Board) Minions() []*Minion { return slices.Clone(b.minions) }

// LivingCount returns the number of alive minions.
func (b Board) LivingCount() int {
	n := 0
	for _, m := range b.minions {
		if m.Alive() {
			n++
		}
	}
	return n
}

// HasLivingAttacker returns true if any alive minion can attack.
func (b Board) HasLivingAttacker() bool {
	for _, m := range b.minions {
		if m.CanAttack() {
			return true
		}
	}
	return false
}

// PickDefender picks a random alive defender.
// Taunt minions are prioritized. Stealth minions cannot be targeted — if only
// Stealth minions remain, returns nil (attacker's turn is skipped).
func (b Board) PickDefender() *Minion {
	var taunt []*Minion
	for _, m := range b.minions {
		if m.Alive() && m.HasKeyword(KeywordTaunt) {
			taunt = append(taunt, m)
		}
	}
	if len(taunt) > 0 {
		return taunt[rand.IntN(len(taunt))] //nolint:gosec // game logic, not crypto
	}

	var targets []*Minion
	for _, m := range b.minions {
		if m.Alive() && !m.HasKeyword(KeywordStealth) {
			targets = append(targets, m)
		}
	}
	if len(targets) > 0 {
		return targets[rand.IntN(len(targets))] //nolint:gosec // game logic, not crypto
	}

	return nil
}

// Clone returns a deep copy of the board (each minion is cloned).
func (b Board) Clone() Board {
	cloned := make([]*Minion, len(b.minions))
	for i, m := range b.minions {
		cloned[i] = m.Clone()
	}
	return Board{minions: cloned}
}

// PlaceMinion inserts a minion at the given position, clamped to valid range.
func (b *Board) PlaceMinion(m *Minion, pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(b.minions) {
		pos = len(b.minions)
	}
	b.minions = append(b.minions[:pos], append([]*Minion{m}, b.minions[pos:]...)...)
}

// RemoveMinion removes and returns the minion at the given index.
// Returns nil if index is out of range.
func (b *Board) RemoveMinion(i int) *Minion {
	if i < 0 || i >= len(b.minions) {
		return nil
	}
	m := b.minions[i]
	b.minions = append(b.minions[:i], b.minions[i+1:]...)
	return m
}

// TribeSnapshot holds a majority tribe and its count.
type TribeSnapshot struct {
	Tribe Tribe
	Count int
}

// MajorityTribe returns the most common non-neutral tribe on the board and its count.
// Returns (TribeNeutral, 0) if no non-neutral tribe has >= 2 minions.
func (b Board) MajorityTribe() (Tribe, int) {
	tribes := make([]Tribe, len(b.minions))
	for i, m := range b.minions {
		tribes[i] = m.Tribe()
	}
	return CalcMajorityTribe(tribes)
}

// CalcMajorityTribe returns the most common non-neutral tribe and its count.
// Returns (TribeNeutral, 0) if no non-neutral tribe has >= 2 entries.
func CalcMajorityTribe(tribes []Tribe) (Tribe, int) {
	counts := make(map[Tribe]int)
	for _, t := range tribes {
		if t == TribeNeutral {
			continue
		}
		counts[t]++
	}

	var bestTribe Tribe
	var bestCount int
	for t, c := range counts {
		if c > bestCount {
			bestTribe = t
			bestCount = c
		}
	}

	if bestCount < 2 {
		if len(counts) > 1 {
			return TribeMixed, 0
		}
		return TribeNeutral, 0
	}
	return bestTribe, bestCount
}

// Reorder reorders the board based on the given index mapping.
func (b *Board) Reorder(order []int) error {
	if len(order) != len(b.minions) {
		return ErrInvalidIndex
	}

	reordered := make([]*Minion, len(b.minions))
	used := make(map[int]struct{}, len(b.minions))

	for i, idx := range order {
		if idx < 0 || idx >= len(b.minions) {
			return ErrInvalidIndex
		}
		if _, ok := used[idx]; ok {
			return ErrInvalidIndex
		}
		reordered[i] = b.minions[idx]
		used[idx] = struct{}{}
	}

	b.minions = reordered
	return nil
}
