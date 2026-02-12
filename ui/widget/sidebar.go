package widget

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/ui"
)

const tierFadeDuration = 2.0

type tierFade struct {
	tier  game.Tier
	timer float64 // seconds remaining
}

// Sidebar displays the player list, hover tooltip, and tier upgrade animations.
type Sidebar struct {
	font      *text.GoTextFace
	hover     int
	tierFades map[string]tierFade
	snap      []client.PlayerEntry // frozen player list
}

func NewSidebar(font *text.GoTextFace) *Sidebar {
	return &Sidebar{
		font:      font,
		hover:     -1,
		tierFades: make(map[string]tierFade),
	}
}

func (s *Sidebar) players() []client.PlayerEntry {
	return s.snap
}

// Update processes tier-fade animations, opponent updates, and hover detection.
func (s *Sidebar) Update(rect ui.Rect, players []client.PlayerEntry, updates []api.OpponentUpdate) {
	if players != nil {
		s.snap = players
	}

	for _, u := range updates {
		s.tierFades[u.PlayerID] = tierFade{tier: u.ShopTier, timer: tierFadeDuration}
	}

	for id, f := range s.tierFades {
		f.timer -= 1.0 / 60.0
		if f.timer <= 0 {
			delete(s.tierFades, id)
		} else {
			s.tierFades[id] = f
		}
	}

	s.updateHover(rect)
}

func (s *Sidebar) updateHover(rect ui.Rect) {
	s.hover = -1
	mx, my := ebiten.CursorPosition()
	if !rect.Contains(mx, my) {
		return
	}
	rowH := rect.H / float64(game.MaxPlayers)
	players := s.players()
	for i := range players {
		row := ui.Rect{X: rect.X, Y: rect.Y + float64(i)*rowH, W: rect.W, H: rowH}
		if row.Contains(mx, my) {
			s.hover = i
			return
		}
	}
}

// Draw renders the sidebar and hover tooltip.
func (s *Sidebar) Draw(screen *ebiten.Image, rect ui.Rect, playerID, opponentID string) {
	s.drawList(screen, rect, playerID, opponentID)
	s.drawTooltip(screen, rect)
}

func (s *Sidebar) drawList(screen *ebiten.Image, rect ui.Rect, playerID, opponentID string) {
	sr := rect.Screen()

	// Background.
	vector.FillRect(screen,
		float32(sr.X), float32(sr.Y),
		float32(sr.W), float32(sr.H),
		color.RGBA{15, 15, 25, 255}, false,
	)

	players := s.players()
	rowH := rect.H / float64(game.MaxPlayers)
	scale := float32(ui.ActiveRes.Scale())

	for i, e := range players {
		rowY := rect.Y + float64(i)*rowH
		row := ui.Rect{X: rect.X, Y: rowY, W: rect.W, H: rowH}
		rowScreen := row.Screen()
		padX := row.W * 0.06

		// Highlight self.
		if e.ID == playerID {
			vector.FillRect(screen,
				float32(rowScreen.X), float32(rowScreen.Y),
				float32(rowScreen.W), float32(rowScreen.H),
				color.RGBA{30, 50, 30, 255}, false,
			)
		}

		// Red border for current combat opponent.
		if e.ID == opponentID {
			vector.StrokeRect(screen,
				float32(rowScreen.X), float32(rowScreen.Y),
				float32(rowScreen.W), float32(rowScreen.H),
				scale*2, color.RGBA{255, 80, 80, 255}, false,
			)
		}

		// Line 1: Name + HP.
		nameClr := color.RGBA{200, 200, 200, 255}
		if e.ID == playerID {
			nameClr = color.RGBA{100, 255, 100, 255}
		}
		ui.DrawText(screen, s.font, fmt.Sprintf("%s  %d HP", e.ID, e.HP), row.X+padX, row.Y+rowH*0.2, nameClr)

		// Line 2: Tier + tribe.
		line2 := fmt.Sprintf("Tier %d", e.ShopTier)
		switch e.MajorityTribe {
		case game.TribeNeutral:
		case game.TribeMixed:
			line2 += " | Mixed"
		default:
			line2 += fmt.Sprintf(" | %s x%d", e.MajorityTribe, e.MajorityCount)
		}
		ui.DrawText(screen, s.font, line2, row.X+padX, row.Y+rowH*0.55, color.RGBA{160, 160, 180, 255})

		// Fading tier upgrade indicator.
		if f, ok := s.tierFades[e.ID]; ok {
			alpha := uint8(255 * (f.timer / tierFadeDuration))
			tierStr := fmt.Sprintf("T%d!", f.tier)
			ui.DrawText(screen, s.font, tierStr, row.X+row.W*0.70, row.Y+rowH*0.35, color.RGBA{255, 215, 0, alpha})
		}

		// Row separator.
		sepY := float32(rowScreen.Bottom())
		vector.StrokeLine(screen,
			float32(rowScreen.X+rowScreen.W*0.05), sepY,
			float32(rowScreen.X+rowScreen.W*0.95), sepY,
			scale, color.RGBA{40, 40, 60, 255}, false,
		)
	}
}

func (s *Sidebar) drawTooltip(screen *ebiten.Image, rect ui.Rect) {
	if s.hover < 0 {
		return
	}
	players := s.players()
	if s.hover >= len(players) {
		return
	}

	e := players[s.hover]
	rowH := rect.H / float64(game.MaxPlayers)
	scale := float32(ui.ActiveRes.Scale())

	// Tooltip positioned to the right of sidebar, aligned with hovered row.
	tipW := rect.W * 1.4
	tipH := rowH * 2.5
	tipX := rect.Right()
	tipY := rect.Y + float64(s.hover)*rowH

	// Clamp to screen bottom.
	if tipY+tipH > float64(ui.BaseHeight) {
		tipY = float64(ui.BaseHeight) - tipH
	}

	tip := ui.Rect{X: tipX, Y: tipY, W: tipW, H: tipH}
	tipScreen := tip.Screen()
	padX := tip.W * 0.06

	// Background.
	vector.FillRect(screen,
		float32(tipScreen.X), float32(tipScreen.Y),
		float32(tipScreen.W), float32(tipScreen.H),
		color.RGBA{25, 25, 40, 255}, false,
	)
	vector.StrokeRect(screen,
		float32(tipScreen.X), float32(tipScreen.Y),
		float32(tipScreen.W), float32(tipScreen.H),
		scale, color.RGBA{60, 60, 90, 255}, false,
	)

	// Header: player name.
	ui.DrawText(screen, s.font, e.ID, tip.X+padX, tip.Y+tip.H*0.08, color.RGBA{220, 220, 220, 255})

	// Tribe info.
	switch e.MajorityTribe {
	case game.TribeNeutral:
	case game.TribeMixed:
		ui.DrawText(screen, s.font, "Mixed", tip.X+padX, tip.Y+tip.H*0.22, color.RGBA{180, 180, 200, 255})
	default:
		tribeStr := fmt.Sprintf("%s x%d", e.MajorityTribe, e.MajorityCount)
		ui.DrawText(screen, s.font, tribeStr, tip.X+padX, tip.Y+tip.H*0.22, color.RGBA{180, 180, 200, 255})
	}

	// Last 3 combat results.
	y := tip.Y + tip.H*0.40
	lineH := tip.H * 0.18
	for _, cr := range e.CombatResults {
		var label string
		var clr color.Color
		switch cr.WinnerID {
		case "":
			label = "Tie vs " + cr.OpponentID
			clr = color.RGBA{140, 140, 140, 255}
		case e.ID:
			label = fmt.Sprintf("Won vs %s (%d dmg)", cr.OpponentID, cr.Damage)
			clr = color.RGBA{80, 220, 80, 255}
		default:
			label = fmt.Sprintf("Lost vs %s (%d dmg)", cr.OpponentID, cr.Damage)
			clr = color.RGBA{220, 80, 80, 255}
		}
		ui.DrawText(screen, s.font, label, tip.X+padX, y, clr)
		y += lineH
	}

	if len(e.CombatResults) == 0 {
		ui.DrawText(screen, s.font, "No fights yet", tip.X+padX, tip.Y+tip.H*0.40, color.RGBA{100, 100, 100, 255})
	}
}
