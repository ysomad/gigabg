package game

import "slices"

// Hand holds cards (minions and spells) with a fixed max capacity.
type Hand struct {
	cards []Card
}

func NewHand() Hand {
	return Hand{cards: make([]Card, 0, maxHandSize)}
}

func (h *Hand) Len() int          { return len(h.cards) }
func (h *Hand) IsFull() bool      { return len(h.cards) >= maxHandSize }
func (h *Hand) CardAt(i int) Card { return h.cards[i] }
func (h *Hand) Cards() []Card     { return slices.Clone(h.cards) }

// Add appends a card to the hand. Does not check capacity.
func (h *Hand) Add(c Card) { h.cards = append(h.cards, c) }

// Remove removes the card at index i and returns it.
func (h *Hand) Remove(i int) Card {
	c := h.cards[i]
	h.cards = append(h.cards[:i], h.cards[i+1:]...)
	return c
}

// RemoveDesc removes cards at the given indices (must be sorted descending).
func (h *Hand) RemoveDesc(indices []int) {
	for _, i := range indices {
		h.cards = slices.Delete(h.cards, i, i+1)
	}
}
