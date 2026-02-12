package game

import (
	"math/rand/v2"
	"slices"
)

const maxCombatIterations = 200

// combatSide holds one player's state during combat simulation.
type combatSide struct {
	player       *Player
	board        Board
	nextAttacker int
}

// nextLivingAttacker advances left-to-right and returns the next minion
// that can attack, or nil if none.
func (s *combatSide) nextLivingAttacker() *Minion {
	n := s.board.Len()
	if n == 0 {
		return nil
	}

	for range n {
		idx := s.nextAttacker % n
		s.nextAttacker = (s.nextAttacker + 1) % n
		if m := s.board.MinionAt(idx); m != nil && m.CanAttack() {
			return m
		}
	}
	return nil
}

// Combat simulates an automated battle between two players.
type Combat struct {
	attacker     *combatSide
	defender     *combatSide
	nextCombatID int
	events       []CombatEvent
	player1ID    string
	player2ID    string
	player1Board Board              // snapshot with combat IDs
	player2Board Board              // snapshot with combat IDs
	poisonKilled map[int]struct{}   // combat IDs killed by poison/venom this attack
}

// NewCombat creates a combat with cloned boards.
// Original player boards are never modified.
func NewCombat(p1, p2 *Player) *Combat {
	c := &Combat{
		nextCombatID: 1,
		player1ID:    p1.ID(),
		player2ID:    p2.ID(),
	}

	side1 := &combatSide{player: p1, board: p1.Board()}
	side2 := &combatSide{player: p2, board: p2.Board()}

	c.assignCombatIDs(&side1.board)
	c.assignCombatIDs(&side2.board)

	// Snapshot boards before combat mutates them.
	c.player1Board = side1.board.Clone()
	c.player2Board = side2.board.Clone()

	if rand.IntN(2) == 1 { //nolint:gosec // game logic, not crypto
		side1, side2 = side2, side1
	}

	c.attacker = side1
	c.defender = side2
	return c
}

func (c *Combat) assignCombatIDs(b *Board) {
	for i := range b.Len() {
		b.MinionAt(i).combatID = c.nextCombatID
		c.nextCombatID++
	}
}

func (c *Combat) emit(e CombatEvent) {
	c.events = append(c.events, e)
}

// Run executes the full combat and returns per-player results.
func (c *Combat) Run() (r1, r2 CombatResult) {
	for range maxCombatIterations {
		if c.attacker.board.LivingCount() == 0 || c.defender.board.LivingCount() == 0 {
			break
		}
		if !c.attacker.board.HasLivingAttacker() && !c.defender.board.HasLivingAttacker() {
			break
		}
		if !c.attacker.board.HasLivingAttacker() {
			c.swapTurns()
			continue
		}

		minion := c.attacker.nextLivingAttacker()
		if minion == nil {
			c.swapTurns()
			continue
		}

		target := c.defender.board.PickDefender()
		if target == nil {
			c.swapTurns()
			continue
		}

		c.attack(minion, target)
		c.removeDeadWithEvents(c.attacker)
		c.removeDeadWithEvents(c.defender)

		// Windfury: attack a second time if still alive and has a target.
		if minion.IsAlive() && minion.HasKeyword(KeywordWindfury) {
			if t2 := c.defender.board.PickDefender(); t2 != nil {
				c.attack(minion, t2)
				c.removeDeadWithEvents(c.attacker)
				c.removeDeadWithEvents(c.defender)
			}
		}

		c.swapTurns()
	}

	return c.results()
}

// attack performs simultaneous damage exchange between two minions.
func (c *Combat) attack(src, dst *Minion) {
	c.emit(CombatEvent{
		Type:     CombatEventAttack,
		SourceID: src.combatID,
		TargetID: dst.combatID,
		OwnerID:  c.attacker.player.ID(),
	})

	c.poisonKilled = make(map[int]struct{})

	hitDst := c.dealDamage(src, dst, src.Attack(), c.defender.player.ID())
	hitSrc := c.dealDamage(dst, src, dst.Attack(), c.attacker.player.ID())

	// Poisonous: kills target on damage (permanent).
	// Venomous: kills target on damage (one-time, keyword consumed).
	c.applyPoison(src, dst, hitDst, c.defender.player.ID())
	c.applyPoison(dst, src, hitSrc, c.attacker.player.ID())
}

// applyPoison checks if src has Poisonous or Venomous and kills dst if damage was dealt.
func (c *Combat) applyPoison(src, dst *Minion, hit bool, ownerID string) {
	if !hit || !dst.IsAlive() {
		return
	}

	isPoisonous := src.HasKeyword(KeywordPoisonous)
	isVenomous := src.HasKeyword(KeywordVenomous)

	if !isPoisonous && !isVenomous {
		return
	}

	// Kill the target and record poison kill.
	dst.TakeDamage(dst.Health())
	c.poisonKilled[dst.combatID] = struct{}{}

	if isVenomous {
		src.RemoveKeyword(KeywordVenomous)
		c.emit(CombatEvent{
			Type:     CombatEventRemoveKeyword,
			TargetID: src.combatID,
			Keyword:  KeywordVenomous,
			OwnerID:  ownerID,
		})
	}
}

// dealDamage applies damage from src to dst. Returns true if damage was dealt.
// Handles Divine Shield: if the target has it, removes the keyword instead.
func (c *Combat) dealDamage(src, dst *Minion, amount int, ownerID string) bool {
	if amount <= 0 {
		return false
	}

	if dst.HasKeyword(KeywordDivineShield) {
		dst.RemoveKeyword(KeywordDivineShield)
		c.emit(CombatEvent{
			Type:     CombatEventRemoveKeyword,
			SourceID: src.combatID,
			TargetID: dst.combatID,
			Keyword:  KeywordDivineShield,
			OwnerID:  ownerID,
		})
		return false
	}

	dst.TakeDamage(amount)
	c.emit(CombatEvent{
		Type:     CombatEventDamage,
		SourceID: src.combatID,
		TargetID: dst.combatID,
		Amount:   amount,
		OwnerID:  ownerID,
	})
	return true
}

// removeDeadWithEvents removes dead minions and emits death events.
func (c *Combat) removeDeadWithEvents(side *combatSide) {
	for i := 0; i < side.board.Len(); i++ {
		m := side.board.MinionAt(i)
		if m.IsAlive() {
			continue
		}
		reason := DeathReasonDamage
		if _, ok := c.poisonKilled[m.combatID]; ok {
			reason = DeathReasonPoison
		}
		c.emit(CombatEvent{
			Type:        CombatEventDeath,
			TargetID:    m.combatID,
			DeathReason: reason,
			OwnerID:     side.player.ID(),
		})
		side.board.RemoveMinion(i)
		if side.nextAttacker > i {
			side.nextAttacker--
		}
		i--
	}
	if n := side.board.Len(); n > 0 {
		side.nextAttacker %= n
	} else {
		side.nextAttacker = 0
	}
}

// swapTurns alternates which side attacks next.
func (c *Combat) swapTurns() {
	c.attacker, c.defender = c.defender, c.attacker
}

// Boards returns snapshots of both players' boards with combat IDs assigned.
func (c *Combat) Boards() (p1Board, p2Board Board) {
	return c.player1Board, c.player2Board
}

// Animation returns the combat animation data for client replay.
func (c *Combat) Animation() CombatAnimation {
	return CombatAnimation{
		Player1ID: c.player1ID,
		Player2ID: c.player2ID,
		Events:    slices.Clone(c.events),
	}
}

// results computes per-player CombatResults from the final board states.
func (c *Combat) results() (CombatResult, CombatResult) {
	p1ID := c.attacker.player.ID()
	p2ID := c.defender.player.ID()

	r1 := CombatResult{OpponentID: p2ID}
	r2 := CombatResult{OpponentID: p1ID}

	alive1 := c.attacker.board.LivingCount()
	alive2 := c.defender.board.LivingCount()

	var winner *combatSide
	switch {
	case alive1 > 0 && alive2 == 0:
		winner = c.attacker
		r1.WinnerID = p1ID
	case alive2 > 0 && alive1 == 0:
		winner = c.defender
		r1.WinnerID = p2ID
	default:
		return r1, r2
	}

	damage := int(winner.player.Shop().Tier())
	for _, m := range winner.board.Minions() {
		if m.IsAlive() {
			damage += int(m.Tier())
		}
	}
	r1.Damage = damage

	r2.WinnerID = r1.WinnerID
	r2.Damage = r1.Damage

	return r1, r2
}
