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
	nextCombatID CombatID
	events       []CombatEvent
	player1      PlayerID
	player2      PlayerID
	player1Board Board                 // snapshot with combat IDs
	player2Board Board                 // snapshot with combat IDs
	poisonKilled map[CombatID]struct{} // killed by poison/venom this attack
}

// NewCombat creates a combat with cloned boards.
// Original player boards are never modified.
func NewCombat(p1, p2 *Player) *Combat {
	c := &Combat{
		nextCombatID: 1,
		player1:      p1.ID(),
		player2:      p2.ID(),
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
	// Stealth is lost when the minion attacks.
	if src.HasKeyword(KeywordStealth) {
		src.RemoveKeyword(KeywordStealth)
		c.emit(RemoveKeywordEvent{
			Target: src.combatID,
			Keyword:  KeywordStealth,
			Owner:    c.attacker.player.ID(),
		})
	}

	c.emit(AttackEvent{
		Source: src.combatID,
		Target: dst.combatID,
		Owner:    c.attacker.player.ID(),
	})

	c.poisonKilled = make(map[CombatID]struct{})
	defender := c.defender.player.ID()

	// Simultaneous damage exchange.
	hitDst := c.dealDamage(src, dst, src.Attack(), defender)
	hitSrc := c.dealDamage(dst, src, dst.Attack(), c.attacker.player.ID())

	// Cleave: damage minions adjacent to the target.
	c.applyCleave(src, dst, defender)

	// Poison/Venom: main target first (may consume Venomous), then cleave targets.
	c.applyPoison(src, dst, hitDst, defender)
	c.applyPoison(dst, src, hitSrc, c.attacker.player.ID())
}

// applyCleave deals attacker's damage to minions adjacent to the target
// and applies poison to each hit target.
func (c *Combat) applyCleave(src, dst *Minion, defender PlayerID) {
	if !src.HasKeyword(KeywordCleave) {
		return
	}

	idx := c.defender.board.IndexOf(dst)
	if idx < 0 {
		return
	}

	for _, adj := range [2]*Minion{
		c.defender.board.MinionAt(idx - 1),
		c.defender.board.MinionAt(idx + 1),
	} {
		if adj == nil {
			continue
		}
		hit := c.dealDamage(src, adj, src.Attack(), defender)
		c.applyPoison(src, adj, hit, defender)                     // src's poison kills adj
		c.applyPoison(adj, src, hit, c.attacker.player.ID()) // adj's poison/venom kills src
	}
}

// applyPoison checks if src has Poisonous or Venomous and kills dst if damage was dealt.
func (c *Combat) applyPoison(src, dst *Minion, hit bool, owner PlayerID) {
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
		c.emit(RemoveKeywordEvent{
			Target: src.combatID,
			Keyword:  KeywordVenomous,
			Owner:    owner,
		})
	}
}

// dealDamage applies damage from src to dst. Returns true if damage was dealt.
// Handles Divine Shield: if the target has it, removes the keyword instead.
func (c *Combat) dealDamage(src, dst *Minion, amount int, owner PlayerID) bool {
	if amount <= 0 {
		return false
	}

	if dst.HasKeyword(KeywordDivineShield) {
		dst.RemoveKeyword(KeywordDivineShield)
		c.emit(RemoveKeywordEvent{
			Source: src.combatID,
			Target: dst.combatID,
			Keyword:  KeywordDivineShield,
			Owner:    owner,
		})
		return false
	}

	dst.TakeDamage(amount)
	c.emit(DamageEvent{
		Source: src.combatID,
		Target: dst.combatID,
		Amount:   amount,
		Owner:    owner,
	})
	return true
}

// removeDeadWithEvents removes dead minions, emits death events,
// and handles Reborn (spawns fresh template minion with 1 HP at the same position).
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
		c.emit(DeathEvent{
			Target:      m.combatID,
			DeathReason: reason,
			Owner:       side.player.ID(),
		})

		hasReborn := m.HasKeyword(KeywordReborn)
		var reborn *Minion
		if hasReborn {
			reborn = NewMinion(m.Template())
			reborn.health = 1
			reborn.RemoveKeyword(KeywordReborn)
			reborn.combatID = c.nextCombatID
			c.nextCombatID++
		}

		side.board.RemoveMinion(i)

		if hasReborn {
			side.board.PlaceMinion(reborn, i)
			c.emit(RebornEvent{
				Target:   reborn.combatID,
				Owner:    side.player.ID(),
				Template: reborn.TemplateID(),
			})
			// Reborn replaced the dead minion in place, no index adjustment.
		} else {
			if side.nextAttacker > i {
				side.nextAttacker--
			}
			i--
		}
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

// Log returns the combat log for client replay.
func (c *Combat) Log() CombatLog {
	return CombatLog{
		Player1: c.player1,
		Player2: c.player2,
		Events:  slices.Clone(c.events),
	}
}

// results computes per-player CombatResults from the final board states.
func (c *Combat) results() (CombatResult, CombatResult) {
	p1 := c.attacker.player.ID()
	p2 := c.defender.player.ID()

	r1 := CombatResult{Opponent: p2}
	r2 := CombatResult{Opponent: p1}

	alive1 := c.attacker.board.LivingCount()
	alive2 := c.defender.board.LivingCount()

	var winnerSide *combatSide
	switch {
	case alive1 > 0 && alive2 == 0:
		winnerSide = c.attacker
		r1.Winner = p1
	case alive2 > 0 && alive1 == 0:
		winnerSide = c.defender
		r1.Winner = p2
	default:
		return r1, r2
	}

	damage := int(winnerSide.player.Shop().Tier())
	for _, m := range winnerSide.board.Minions() {
		if m.IsAlive() {
			damage += int(m.Tier())
		}
	}
	r1.Damage = damage

	r2.Winner = r1.Winner
	r2.Damage = r1.Damage

	return r1, r2
}
