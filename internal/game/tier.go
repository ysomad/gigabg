package game

type Tier uint8

const (
	Tier1 Tier = 1
	Tier2 Tier = 2
	Tier3 Tier = 3
	Tier4 Tier = 4
	Tier5 Tier = 5
	Tier6 Tier = 6
)

func (t Tier) IsValid() bool {
	return t >= Tier1 || t <= Tier6
}
