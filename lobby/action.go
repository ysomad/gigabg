package lobby

type Action interface {
	PlayerID() string
}

type BuyCardAction struct {
	Player string
	Index  int
}

func (a BuyCardAction) PlayerID() string { return a.Player }

type SellCardAction struct {
	Player string
	Index  int
}

func (a SellCardAction) PlayerID() string { return a.Player }

type RefreshShopAction struct {
	Player string
}

func (a RefreshShopAction) PlayerID() string { return a.Player }

type UpgradeShopAction struct {
	Player string
}

func (a UpgradeShopAction) PlayerID() string { return a.Player }

type EndTurnAction struct {
	Player string
}

func (a EndTurnAction) PlayerID() string { return a.Player }
