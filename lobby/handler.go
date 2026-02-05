package lobby

import (
	"github.com/ysomad/gigabg/errorsx"
	"github.com/ysomad/gigabg/game"
)

const (
	ErrUnknownAction  errorsx.Error = "unknown action type"
	ErrPlayerNotFound errorsx.Error = "player not found"
	ErrNotAMinion     errorsx.Error = "card is not a minion"
)

// Execute processes an action and mutates lobby state.
func (l *Lobby) Execute(action Action) error {
	player := l.Player(action.PlayerID())
	if player == nil {
		return ErrPlayerNotFound
	}

	switch a := action.(type) {
	case BuyCardAction:
		return l.handleBuy(player, a.Index)
	case SellCardAction:
		return l.handleSell(player, a.Index)
	case RefreshShopAction:
		return l.handleRefresh(player)
	case UpgradeShopAction:
		return l.handleUpgrade(player)
	case EndTurnAction:
		return l.handleEndTurn()
	default:
		return ErrUnknownAction
	}
}

func (l *Lobby) handleBuy(p *game.Player, index int) error {
	if index < 0 || index >= len(p.Shop) {
		return game.ErrInvalidIndex
	}
	if p.Gold < game.BuyCost {
		return game.ErrNotEnoughGold
	}
	if len(p.Board) >= game.MaxBoardSize {
		return game.ErrBoardFull
	}

	card := p.Shop[index]
	minion, ok := card.(*game.Minion)
	if !ok {
		return ErrNotAMinion
	}

	p.Gold -= game.BuyCost
	p.Shop = append(p.Shop[:index], p.Shop[index+1:]...)
	p.Board = append(p.Board, minion)

	return nil
}

func (l *Lobby) handleSell(p *game.Player, index int) error {
	if index < 0 || index >= len(p.Board) {
		return game.ErrInvalidIndex
	}

	p.Board = append(p.Board[:index], p.Board[index+1:]...)
	p.Gold += game.SellValue
	if p.Gold > game.MaxGold {
		p.Gold = game.MaxGold
	}

	return nil
}

func (l *Lobby) handleRefresh(p *game.Player) error {
	if p.Gold < game.RefreshCost {
		return game.ErrNotEnoughGold
	}

	p.Gold -= game.RefreshCost
	p.Shop = l.pool.Roll(p.ShopTier, nil, p.ShopSize())

	return nil
}

func (l *Lobby) handleUpgrade(p *game.Player) error {
	if p.ShopTier >= 6 {
		return game.ErrMaxTier
	}

	cost := p.UpgradeCost()
	if p.Gold < cost {
		return game.ErrNotEnoughGold
	}

	p.Gold -= cost
	p.ShopTier++

	return nil
}

func (l *Lobby) handleEndTurn() error {
	l.nextTurn()
	return nil
}
