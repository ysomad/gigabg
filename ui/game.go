package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/message"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
)

var ColorBackground = color.RGBA{20, 20, 30, 255}

// GameScene renders the game UI.
type GameScene struct {
	client *client.RemoteClient
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

func NewGameScene(c *client.RemoteClient, cards *cards.Cards, font *text.GoTextFace) *GameScene {
	return &GameScene{
		client: c,
		cards:  cards,
		font:   font,
	}
}

const (
	cardWidth  = 130
	cardHeight = 120
	cardGap    = 20
	shopY      = 170
	boardY     = 340
	handY      = 510

	refreshBtnX = 400
	refreshBtnY = 130
	refreshBtnW = 120
	refreshBtnH = 30

	upgradeBtnX = 530
	upgradeBtnY = 130
	upgradeBtnW = 150
	upgradeBtnH = 30
)

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

	// Start drag
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// From board (use local order)
		if y >= boardY && y <= boardY+cardHeight {
			for i := range g.boardOrder {
				cardX := 50 + i*(cardWidth+cardGap)
				if x >= cardX && x <= cardX+cardWidth {
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
		if y >= handY && y <= handY+cardHeight {
			hand := g.client.Hand()
			for i, c := range hand {
				cardX := 50 + i*(cardWidth+cardGap)
				if x >= cardX && x <= cardX+cardWidth {
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
		if y >= boardY-50 && y <= boardY+cardHeight+50 {
			if g.dragFromBoard {
				// Reorder locally
				pos := g.getDropPosition(x)
				if pos != g.dragIndex && pos >= 0 && pos <= len(g.boardOrder) {
					// Remove from old position and insert at new
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
				// Board order will be reset when board size changes
			}
		}
		g.dragging = false
		return nil
	}

	// Click shop to buy or refresh
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		// Refresh shop button
		if x >= refreshBtnX && x <= refreshBtnX+refreshBtnW &&
			y >= refreshBtnY && y <= refreshBtnY+refreshBtnH {
			g.client.RefreshShop()
			return nil
		}

		// Upgrade shop button
		if x >= upgradeBtnX && x <= upgradeBtnX+upgradeBtnW &&
			y >= upgradeBtnY && y <= upgradeBtnY+upgradeBtnH {
			g.client.UpgradeShop()
			return nil
		}

		// Buy from shop
		if y >= shopY && y <= shopY+cardHeight {
			shop := g.client.Shop()
			for i := range shop {
				cardX := 50 + i*(cardWidth+cardGap)
				if x >= cardX && x <= cardX+cardWidth {
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

	for i := range board {
		cardCenterX := 50 + i*(cardWidth+cardGap) + cardWidth/2
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
	drawText(screen, g.font, "Connecting...", ScreenWidth/2-50, ScreenHeight/2, color.White)
}

func (g *GameScene) drawWaiting(screen *ebiten.Image, playerCount int) {
	myID := g.client.PlayerID()

	header := fmt.Sprintf("You are Player %s | Waiting for players... %d/%d", myID, playerCount, game.MaxPlayers)
	drawText(screen, g.font, header, 50, 50, color.RGBA{200, 200, 200, 255})

	vector.StrokeLine(screen, 40, 80, ScreenWidth-40, 80, 1, color.RGBA{60, 60, 80, 255}, false)
}

func (g *GameScene) drawRecruit(screen *ebiten.Image) {
	turn := g.client.Turn()
	remaining := g.timeRemaining()

	header := fmt.Sprintf("Turn %d | RECRUIT PHASE", turn)
	drawText(screen, g.font, header, 50, 50, color.RGBA{100, 200, 100, 255})

	timer := fmt.Sprintf("%d:%02d", remaining/60, remaining%60)
	drawText(screen, g.font, timer, ScreenWidth-100, 50, color.RGBA{255, 255, 255, 255})

	vector.StrokeLine(screen, 40, 80, ScreenWidth-40, 80, 1, color.RGBA{60, 60, 80, 255}, false)

	// Player stats
	if p := g.client.Player(); p != nil {
		stats := fmt.Sprintf("HP: %d | Gold: %d/%d | Tier: %d | Upgrade: %d gold",
			p.HP, p.Gold, p.MaxGold, p.ShopTier, p.UpgradeCost)
		drawText(screen, g.font, stats, 50, 100, color.RGBA{255, 215, 0, 255})
	}

	// Shop
	drawText(screen, g.font, "SHOP (click to buy for 3 gold)", 50, 140, color.RGBA{150, 150, 150, 255})

	// Refresh button
	vector.FillRect(screen, refreshBtnX, refreshBtnY, refreshBtnW, refreshBtnH, color.RGBA{60, 60, 90, 255}, false)
	vector.StrokeRect(
		screen,
		refreshBtnX,
		refreshBtnY,
		refreshBtnW,
		refreshBtnH,
		1,
		color.RGBA{100, 100, 140, 255},
		false,
	)
	drawText(
		screen,
		g.font,
		"Refresh (1g)",
		float64(refreshBtnX)+8,
		float64(refreshBtnY)+7,
		color.RGBA{200, 200, 255, 255},
	)

	// Upgrade button
	if p := g.client.Player(); p != nil {
		vector.FillRect(screen, upgradeBtnX, upgradeBtnY, upgradeBtnW, upgradeBtnH, color.RGBA{60, 90, 60, 255}, false)
		vector.StrokeRect(
			screen,
			upgradeBtnX,
			upgradeBtnY,
			upgradeBtnW,
			upgradeBtnH,
			1,
			color.RGBA{100, 140, 100, 255},
			false,
		)
		label := fmt.Sprintf("Upgrade (%dg)", p.UpgradeCost)
		drawText(screen, g.font, label, float64(upgradeBtnX)+8, float64(upgradeBtnY)+7, color.RGBA{200, 255, 200, 255})
	}
	shop := g.client.Shop()
	for i, c := range shop {
		x := float64(50 + i*(cardWidth+cardGap))
		g.drawCard(screen, c, x, shopY)
	}

	// Board (use local order)
	drawText(screen, g.font, "BOARD", 50, 310, color.RGBA{150, 150, 150, 255})
	board := g.client.Board()
	for i, serverIdx := range g.boardOrder {
		if g.dragging && g.dragFromBoard && i == g.dragIndex {
			continue // skip dragged card
		}
		if serverIdx >= 0 && serverIdx < len(board) {
			x := float64(50 + i*(cardWidth+cardGap))
			g.drawCard(screen, board[serverIdx], x, boardY)
		}
	}

	// Hand
	drawText(screen, g.font, "HAND (drag to board)", 50, 480, color.RGBA{150, 150, 150, 255})
	hand := g.client.Hand()
	for i, c := range hand {
		if g.dragging && !g.dragFromBoard && i == g.dragIndex {
			continue // skip dragged card
		}
		x := float64(50 + i*(cardWidth+cardGap))
		g.drawCard(screen, c, x, handY)
	}

	// Draw dragged card at cursor
	if g.dragging {
		var c message.Card
		if g.dragFromBoard {
			serverIdx := g.boardOrder[g.dragIndex]
			if serverIdx >= 0 && serverIdx < len(board) {
				c = board[serverIdx]
			}
		} else {
			c = hand[g.dragIndex]
		}
		g.drawCard(screen, c, float64(g.dragCurrentX-cardWidth/2), float64(g.dragCurrentY-cardHeight/2))
	}

	// Discover overlay
	if discover := g.client.Discover(); discover != nil {
		g.drawDiscoverOverlay(screen, discover)
	}
}

func (g *GameScene) drawDiscoverOverlay(screen *ebiten.Image, discover *message.DiscoverOffer) {
	// Semi-transparent background
	vector.FillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 160}, false)

	drawText(
		screen,
		g.font,
		"DISCOVER — Pick a card",
		float64(ScreenWidth/2-80),
		float64(ScreenHeight/2-cardHeight/2-30),
		color.RGBA{255, 215, 0, 255},
	)

	startX := discoverStartX(len(discover.Cards))
	y := float64(ScreenHeight/2 - cardHeight/2)
	for i, c := range discover.Cards {
		x := float64(startX + i*(cardWidth+cardGap))
		g.drawCard(screen, c, x, y)
	}
}

func (g *GameScene) handleDiscoverClick(discover *message.DiscoverOffer, x, y int) {
	discoverY := ScreenHeight/2 - cardHeight/2
	for i := range discover.Cards {
		cardX := discoverStartX(len(discover.Cards)) + i*(cardWidth+cardGap)
		if x >= cardX && x <= cardX+cardWidth && y >= discoverY && y <= discoverY+cardHeight {
			g.client.DiscoverPick(i)
			return
		}
	}
}

func discoverStartX(count int) int {
	totalW := count*(cardWidth+cardGap) - cardGap
	return (ScreenWidth - totalW) / 2
}

func (g *GameScene) drawCard(screen *ebiten.Image, c message.Card, x, y float64) {
	t := g.cards.ByTemplateID(c.TemplateID)
	isSpell := t != nil && t.IsSpell()

	// Card background
	bg := color.RGBA{40, 40, 60, 255}
	if isSpell {
		bg = color.RGBA{80, 40, 100, 255}
	}
	vector.FillRect(screen, float32(x), float32(y), cardWidth, cardHeight, bg, false)

	// Border — golden for golden cards, purple for spells, default otherwise
	border := color.RGBA{80, 80, 100, 255}
	borderWidth := float32(2)
	if c.Golden {
		border = color.RGBA{255, 215, 0, 255}
		borderWidth = 3
	} else if isSpell {
		border = color.RGBA{140, 80, 180, 255}
	}
	vector.StrokeRect(screen, float32(x), float32(y), cardWidth, cardHeight, borderWidth, border, false)

	// Get template for name, description, tribe
	name := c.TemplateID
	desc := ""
	tribe := ""
	if t != nil {
		name = t.Name
		desc = t.Description
		tribe = t.Tribe.String()
	}

	// Name at top left
	drawText(screen, g.font, name, x+5, y+5, color.White)

	// Tier at top right
	if t != nil && !isSpell && t.Tier.IsValid() {
		drawText(screen, g.font, fmt.Sprintf("T%d", t.Tier), x+cardWidth-30, y+5, color.RGBA{180, 180, 180, 255})
	}

	// Description in center
	drawText(screen, g.font, desc, x+5, y+40, color.RGBA{180, 180, 180, 255})

	if isSpell {
		// Spell label at bottom center
		drawText(screen, g.font, "SPELL", x+cardWidth/2-20, y+cardHeight-18, color.RGBA{200, 150, 255, 255})
	} else {
		// Tribe in center bottom
		drawText(screen, g.font, tribe, x+cardWidth/2-20, y+cardHeight-35, color.RGBA{150, 150, 200, 255})

		// AP in left bottom (yellow)
		drawText(screen, g.font, fmt.Sprintf("%d", c.Attack), x+5, y+cardHeight-18, color.RGBA{255, 215, 0, 255})

		// HP in right bottom (red)
		drawText(
			screen,
			g.font,
			fmt.Sprintf("%d", c.Health),
			x+cardWidth-20,
			y+cardHeight-18,
			color.RGBA{255, 80, 80, 255},
		)
	}
}

func (g *GameScene) drawCombat(screen *ebiten.Image) {
	turn := g.client.Turn()
	remaining := g.timeRemaining()

	header := fmt.Sprintf("Turn %d | COMBAT PHASE", turn)
	drawText(screen, g.font, header, 50, 50, color.RGBA{255, 100, 100, 255})

	timer := fmt.Sprintf("%d:%02d", remaining/60, remaining%60)
	drawText(screen, g.font, timer, ScreenWidth-100, 50, color.RGBA{255, 255, 255, 255})

	vector.StrokeLine(screen, 40, 80, ScreenWidth-40, 80, 1, color.RGBA{60, 60, 80, 255}, false)
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

	x := float64(50)
	y := float64(ScreenHeight - 30)

	for _, p := range players {
		label := fmt.Sprintf("%s (%d)", p.ID, p.HP)

		clr := color.RGBA{200, 200, 200, 255}
		if p.ID == myID {
			clr = color.RGBA{100, 255, 100, 255}
		}

		drawText(screen, g.font, label, x, y, clr)
		x += 80
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
