package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/cards"
)

var ColorBackground = color.RGBA{20, 20, 30, 255}

// Base layout constants (designed for 1280x720, scaled at runtime).
const (
	baseCardWidth  = 130
	baseCardHeight = 120
	baseCardGap    = 20
	baseShopY      = 170
	baseBoardY     = 340
	baseHandY      = 510

	baseRefreshBtnX = 400
	baseRefreshBtnY = 130
	baseRefreshBtnW = 120
	baseRefreshBtnH = 30

	baseUpgradeBtnX = 530
	baseUpgradeBtnY = 130
	baseUpgradeBtnW = 150
	baseUpgradeBtnH = 30
)

// Scaled layout helpers.
func sc(base int) float64    { return float64(base) * ActiveRes.Scale() }
func scf(base float64) float64 { return base * ActiveRes.Scale() }
func sci(base int) int        { return int(sc(base)) }

// GameScene renders the game UI.
type GameScene struct {
	client *client.Client
	cards  *cards.Cards
	font   *text.GoTextFace

	// Drag state
	dragging      bool
	dragIndex     int
	dragFromBoard bool
	dragCurrentX  int
	dragCurrentY  int

	// Local board order (indices into server board)
	boardOrder []int
	lastPhase  game.Phase
}

func NewGameScene(c *client.Client, cards *cards.Cards, font *text.GoTextFace) *GameScene {
	return &GameScene{
		client: c,
		cards:  cards,
		font:   font,
	}
}

func (g *GameScene) Update() error {
	phase := g.client.Phase()

	// Sync board order when transitioning from recruit to combat
	if g.lastPhase == game.PhaseRecruit && phase == game.PhaseCombat {
		if len(g.boardOrder) > 0 {
			g.client.SyncBoard(g.boardOrder)
		}
	}
	g.lastPhase = phase

	// Reset board order when board size changes (new card placed or removed)
	board := g.client.Board()
	if len(g.boardOrder) != len(board) {
		g.boardOrder = make([]int, len(board))
		for i := range g.boardOrder {
			g.boardOrder[i] = i
		}
	}

	if phase != game.PhaseRecruit {
		return nil
	}

	x, y := ebiten.CursorPosition()

	// Handle discover overlay (blocks all other input)
	if discover := g.client.Discover(); discover != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleDiscoverClick(discover, x, y)
		}
		return nil
	}

	cardW := sci(baseCardWidth)
	cardH := sci(baseCardHeight)
	cardStep := sci(baseCardWidth + baseCardGap)
	margin := sci(50)
	boardYs := sci(baseBoardY)
	handYs := sci(baseHandY)
	shopYs := sci(baseShopY)

	// Start drag
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// From board (use local order)
		if y >= boardYs && y <= boardYs+cardH {
			for i := range g.boardOrder {
				cardX := margin + i*cardStep
				if x >= cardX && x <= cardX+cardW {
					g.dragging = true
					g.dragIndex = i
					g.dragFromBoard = true
					g.dragCurrentX = x
					g.dragCurrentY = y
					return nil
				}
			}
		}
		// From hand — click spell to play, drag minion
		if y >= handYs && y <= handYs+cardH {
			hand := g.client.Hand()
			for i, c := range hand {
				cardX := margin + i*cardStep
				if x >= cardX && x <= cardX+cardW {
					if t := g.cards.ByTemplateID(c.TemplateID); t != nil && t.IsSpell() {
						g.client.PlaySpell(i)
						return nil
					}
					g.dragging = true
					g.dragIndex = i
					g.dragFromBoard = false
					g.dragCurrentX = x
					g.dragCurrentY = y
					return nil
				}
			}
		}
	}

	// Update drag position
	if g.dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.dragCurrentX = x
		g.dragCurrentY = y
		return nil
	}

	// End drag
	if g.dragging && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		dropPad := sci(50)
		if y >= boardYs-dropPad && y <= boardYs+cardH+dropPad {
			if g.dragFromBoard {
				// Reorder locally
				pos := g.getDropPosition(x)
				if pos != g.dragIndex && pos >= 0 && pos <= len(g.boardOrder) {
					val := g.boardOrder[g.dragIndex]
					g.boardOrder = append(g.boardOrder[:g.dragIndex], g.boardOrder[g.dragIndex+1:]...)
					if pos > g.dragIndex {
						pos--
					}
					g.boardOrder = append(g.boardOrder[:pos], append([]int{val}, g.boardOrder[pos:]...)...)
				}
			} else {
				// Place from hand to server
				pos := g.getDropPosition(x)
				g.client.PlaceMinion(g.dragIndex, pos)
			}
		}
		g.dragging = false
		return nil
	}

	// Click shop to buy or refresh
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		// Refresh shop button
		rbx := sci(baseRefreshBtnX)
		rby := sci(baseRefreshBtnY)
		rbw := sci(baseRefreshBtnW)
		rbh := sci(baseRefreshBtnH)
		if x >= rbx && x <= rbx+rbw && y >= rby && y <= rby+rbh {
			g.client.RefreshShop()
			return nil
		}

		// Upgrade shop button
		ubx := sci(baseUpgradeBtnX)
		uby := sci(baseUpgradeBtnY)
		ubw := sci(baseUpgradeBtnW)
		ubh := sci(baseUpgradeBtnH)
		if x >= ubx && x <= ubx+ubw && y >= uby && y <= uby+ubh {
			g.client.UpgradeShop()
			return nil
		}

		// Buy from shop
		if y >= shopYs && y <= shopYs+cardH {
			shop := g.client.Shop()
			for i := range shop {
				cardX := margin + i*cardStep
				if x >= cardX && x <= cardX+cardW {
					g.client.BuyCard(i)
					return nil
				}
			}
		}
	}

	return nil
}

func (g *GameScene) getDropPosition(x int) int {
	board := g.client.Board()
	if len(board) == 0 {
		return 0
	}

	margin := sci(50)
	cardStep := sci(baseCardWidth + baseCardGap)
	halfCard := sci(baseCardWidth) / 2

	for i := range board {
		cardCenterX := margin + i*cardStep + halfCard
		if x < cardCenterX {
			return i
		}
	}
	return len(board)
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBackground)

	playerCount := len(g.client.Players())
	if playerCount == 0 {
		g.drawConnecting(screen)
		return
	}

	phase := g.client.Phase()

	switch phase {
	case game.PhaseWaiting:
		g.drawWaiting(screen, playerCount)
	case game.PhaseRecruit:
		g.drawRecruit(screen)
	case game.PhaseCombat:
		g.drawCombat(screen)
	}

	g.drawPlayers(screen)
}

func (g *GameScene) drawConnecting(screen *ebiten.Image) {
	w := float64(ActiveRes.Width)
	h := float64(ActiveRes.Height)
	drawText(screen, g.font, "Connecting...", w/2-scf(50), h/2, color.White)
}

func (g *GameScene) drawWaiting(screen *ebiten.Image, playerCount int) {
	myID := g.client.PlayerID()

	header := fmt.Sprintf("You are Player %s | Waiting for players... %d/%d", myID, playerCount, game.MaxPlayers)
	drawText(screen, g.font, header, sc(50), sc(50), color.RGBA{200, 200, 200, 255})

	w := float64(ActiveRes.Width)
	vector.StrokeLine(screen, float32(sc(40)), float32(sc(80)), float32(w-sc(40)), float32(sc(80)), 1, color.RGBA{60, 60, 80, 255}, false)
}

func (g *GameScene) drawRecruit(screen *ebiten.Image) {
	turn := g.client.Turn()
	remaining := g.timeRemaining()
	w := float64(ActiveRes.Width)

	header := fmt.Sprintf("Turn %d | RECRUIT PHASE", turn)
	drawText(screen, g.font, header, sc(50), sc(50), color.RGBA{100, 200, 100, 255})

	timer := fmt.Sprintf("%d:%02d", remaining/60, remaining%60)
	drawText(screen, g.font, timer, w-sc(100), sc(50), color.RGBA{255, 255, 255, 255})

	vector.StrokeLine(screen, float32(sc(40)), float32(sc(80)), float32(w-sc(40)), float32(sc(80)), 1, color.RGBA{60, 60, 80, 255}, false)

	// Player stats
	if p := g.client.Player(); p != nil {
		stats := fmt.Sprintf("HP: %d | Gold: %d/%d | Tier: %d | Upgrade: %d gold",
			p.HP, p.Gold, p.MaxGold, p.ShopTier, p.UpgradeCost)
		drawText(screen, g.font, stats, sc(50), sc(100), color.RGBA{255, 215, 0, 255})
	}

	// Shop
	drawText(screen, g.font, "SHOP (click to buy for 3 gold)", sc(50), sc(140), color.RGBA{150, 150, 150, 255})

	// Refresh button
	rbx := float32(sc(baseRefreshBtnX))
	rby := float32(sc(baseRefreshBtnY))
	rbw := float32(sc(baseRefreshBtnW))
	rbh := float32(sc(baseRefreshBtnH))
	vector.FillRect(screen, rbx, rby, rbw, rbh, color.RGBA{60, 60, 90, 255}, false)
	vector.StrokeRect(screen, rbx, rby, rbw, rbh, 1, color.RGBA{100, 100, 140, 255}, false)
	drawText(screen, g.font, "Refresh (1g)", float64(rbx)+scf(8), float64(rby)+scf(7), color.RGBA{200, 200, 255, 255})

	// Upgrade button
	if p := g.client.Player(); p != nil {
		ubx := float32(sc(baseUpgradeBtnX))
		uby := float32(sc(baseUpgradeBtnY))
		ubw := float32(sc(baseUpgradeBtnW))
		ubh := float32(sc(baseUpgradeBtnH))
		vector.FillRect(screen, ubx, uby, ubw, ubh, color.RGBA{60, 90, 60, 255}, false)
		vector.StrokeRect(screen, ubx, uby, ubw, ubh, 1, color.RGBA{100, 140, 100, 255}, false)
		label := fmt.Sprintf("Upgrade (%dg)", p.UpgradeCost)
		drawText(screen, g.font, label, float64(ubx)+scf(8), float64(uby)+scf(7), color.RGBA{200, 255, 200, 255})
	}

	shop := g.client.Shop()
	for i, c := range shop {
		x := sc(50) + float64(i)*sc(baseCardWidth+baseCardGap)
		g.drawCard(screen, c, x, sc(baseShopY))
	}

	// Board (use local order)
	drawText(screen, g.font, "BOARD", sc(50), sc(310), color.RGBA{150, 150, 150, 255})
	board := g.client.Board()
	for i, serverIdx := range g.boardOrder {
		if g.dragging && g.dragFromBoard && i == g.dragIndex {
			continue
		}
		if serverIdx >= 0 && serverIdx < len(board) {
			x := sc(50) + float64(i)*sc(baseCardWidth+baseCardGap)
			g.drawCard(screen, board[serverIdx], x, sc(baseBoardY))
		}
	}

	// Hand
	drawText(screen, g.font, "HAND (drag to board)", sc(50), sc(480), color.RGBA{150, 150, 150, 255})
	hand := g.client.Hand()
	for i, c := range hand {
		if g.dragging && !g.dragFromBoard && i == g.dragIndex {
			continue
		}
		x := sc(50) + float64(i)*sc(baseCardWidth+baseCardGap)
		g.drawCard(screen, c, x, sc(baseHandY))
	}

	// Draw dragged card at cursor
	if g.dragging {
		var c api.Card
		if g.dragFromBoard {
			serverIdx := g.boardOrder[g.dragIndex]
			if serverIdx >= 0 && serverIdx < len(board) {
				c = board[serverIdx]
			}
		} else {
			c = hand[g.dragIndex]
		}
		g.drawCard(screen, c, float64(g.dragCurrentX)-sc(baseCardWidth)/2, float64(g.dragCurrentY)-sc(baseCardHeight)/2)
	}

	// Discover overlay
	if discover := g.client.Discover(); discover != nil {
		g.drawDiscoverOverlay(screen, discover)
	}
}

func (g *GameScene) drawDiscoverOverlay(screen *ebiten.Image, discover *api.DiscoverOffer) {
	w := float32(ActiveRes.Width)
	h := float32(ActiveRes.Height)
	vector.FillRect(screen, 0, 0, w, h, color.RGBA{0, 0, 0, 160}, false)

	cardH := sc(baseCardHeight)
	drawText(
		screen,
		g.font,
		"DISCOVER — Pick a card",
		float64(ActiveRes.Width)/2-scf(80),
		float64(ActiveRes.Height)/2-cardH/2-scf(30),
		color.RGBA{255, 215, 0, 255},
	)

	startX := g.discoverStartX(len(discover.Cards))
	y := float64(ActiveRes.Height)/2 - cardH/2
	for i, c := range discover.Cards {
		x := startX + float64(i)*sc(baseCardWidth+baseCardGap)
		g.drawCard(screen, c, x, y)
	}
}

func (g *GameScene) handleDiscoverClick(discover *api.DiscoverOffer, x, y int) {
	cardH := sci(baseCardHeight)
	cardW := sci(baseCardWidth)
	cardStep := sci(baseCardWidth + baseCardGap)
	discoverY := ActiveRes.Height/2 - cardH/2
	startX := int(g.discoverStartX(len(discover.Cards)))

	for i := range discover.Cards {
		cardX := startX + i*cardStep
		if x >= cardX && x <= cardX+cardW && y >= discoverY && y <= discoverY+cardH {
			g.client.DiscoverPick(i)
			return
		}
	}
}

func (g *GameScene) discoverStartX(count int) float64 {
	totalW := sc(count*(baseCardWidth+baseCardGap) - baseCardGap)
	return (float64(ActiveRes.Width) - totalW) / 2
}

func (g *GameScene) drawCard(screen *ebiten.Image, c api.Card, x, y float64) {
	t := g.cards.ByTemplateID(c.TemplateID)
	isSpell := t != nil && t.IsSpell()

	cw := float32(sc(baseCardWidth))
	ch := float32(sc(baseCardHeight))

	// Card background
	bg := color.RGBA{40, 40, 60, 255}
	if isSpell {
		bg = color.RGBA{80, 40, 100, 255}
	}
	vector.FillRect(screen, float32(x), float32(y), cw, ch, bg, false)

	// Border
	border := color.RGBA{80, 80, 100, 255}
	borderWidth := float32(2)
	if c.Golden {
		border = color.RGBA{255, 215, 0, 255}
		borderWidth = 3
	} else if isSpell {
		border = color.RGBA{140, 80, 180, 255}
	}
	vector.StrokeRect(screen, float32(x), float32(y), cw, ch, borderWidth, border, false)

	name := c.TemplateID
	desc := ""
	tribe := ""
	if t != nil {
		name = t.Name
		desc = t.Description
		tribe = t.Tribe.String()
	}

	// Name at top left
	drawText(screen, g.font, name, x+scf(5), y+scf(5), color.White)

	// Tier at top right
	if t != nil && !isSpell && t.Tier.IsValid() {
		drawText(screen, g.font, fmt.Sprintf("T%d", t.Tier), x+sc(baseCardWidth)-scf(30), y+scf(5), color.RGBA{180, 180, 180, 255})
	}

	// Description in center
	drawText(screen, g.font, desc, x+scf(5), y+scf(40), color.RGBA{180, 180, 180, 255})

	if isSpell {
		drawText(screen, g.font, "SPELL", x+sc(baseCardWidth)/2-scf(20), y+sc(baseCardHeight)-scf(18), color.RGBA{200, 150, 255, 255})
	} else {
		// Tribe
		drawText(screen, g.font, tribe, x+sc(baseCardWidth)/2-scf(20), y+sc(baseCardHeight)-scf(35), color.RGBA{150, 150, 200, 255})

		// AP bottom-left
		drawText(screen, g.font, fmt.Sprintf("%d", c.Attack), x+scf(5), y+sc(baseCardHeight)-scf(18), color.RGBA{255, 215, 0, 255})

		// HP bottom-right
		drawText(screen, g.font, fmt.Sprintf("%d", c.Health), x+sc(baseCardWidth)-scf(20), y+sc(baseCardHeight)-scf(18), color.RGBA{255, 80, 80, 255})
	}
}

func (g *GameScene) drawCombat(screen *ebiten.Image) {
	turn := g.client.Turn()
	remaining := g.timeRemaining()
	w := float64(ActiveRes.Width)

	header := fmt.Sprintf("Turn %d | COMBAT PHASE", turn)
	drawText(screen, g.font, header, sc(50), sc(50), color.RGBA{255, 100, 100, 255})

	timer := fmt.Sprintf("%d:%02d", remaining/60, remaining%60)
	drawText(screen, g.font, timer, w-sc(100), sc(50), color.RGBA{255, 255, 255, 255})

	vector.StrokeLine(screen, float32(sc(40)), float32(sc(80)), float32(w-sc(40)), float32(sc(80)), 1, color.RGBA{60, 60, 80, 255}, false)
}

func (g *GameScene) timeRemaining() int64 {
	endTime := g.client.PhaseEndTimestamp()
	now := time.Now().Unix()
	remaining := endTime - now
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (g *GameScene) drawPlayers(screen *ebiten.Image) {
	players := g.client.Players()
	myID := g.client.PlayerID()

	x := sc(50)
	y := float64(ActiveRes.Height) - sc(30)

	for _, p := range players {
		label := fmt.Sprintf("%s (%d)", p.ID, p.HP)

		clr := color.RGBA{200, 200, 200, 255}
		if p.ID == myID {
			clr = color.RGBA{100, 255, 100, 255}
		}

		drawText(screen, g.font, label, x, y, clr)
		x += sc(80)
	}
}

func drawText(screen *ebiten.Image, font *text.GoTextFace, str string, x, y float64, clr color.Color) {
	if font == nil {
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, font, op)
}

func DrawTextAt(screen *ebiten.Image, font *text.GoTextFace, str string, x, y float64) {
	drawText(screen, font, str, x, y, color.White)
}
