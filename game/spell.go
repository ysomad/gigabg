package game

// Spell is a runtime spell instance wrapping a CardTemplate.
type Spell struct {
	template *CardTemplate
}

// NewSpell creates a spell from a card template.
func NewSpell(t *CardTemplate) *Spell {
	return &Spell{template: t}
}

func (s *Spell) Template() *CardTemplate { return s.template }
func (s *Spell) TemplateID() string      { return s.template.ID }
func (s *Spell) Name() string            { return s.template.Name }
func (s *Spell) Tier() Tier              { return s.template.Tier }
func (s *Spell) IsSpell() bool  { return true }
func (s *Spell) IsMinion() bool { return false }
func (s *Spell) IsGolden() bool { return false }
