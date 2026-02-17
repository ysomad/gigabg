package ui

import "github.com/hajimehoshi/ebiten/v2"

// Scene is the interface for all UI scenes.
type Scene interface {
	Update(res Resolution) error
	Draw(screen *ebiten.Image, res Resolution)
	OnEnter()
	OnExit()
}

// SceneManager manages the active scene.
type SceneManager struct {
	current Scene
}

func (sm *SceneManager) Switch(s Scene) {
	if sm.current != nil {
		sm.current.OnExit()
	}
	sm.current = s
	if sm.current != nil {
		sm.current.OnEnter()
	}
}

func (sm *SceneManager) Update(res Resolution) error {
	if sm.current == nil {
		return nil
	}
	return sm.current.Update(res)
}

func (sm *SceneManager) Draw(screen *ebiten.Image, res Resolution) {
	if sm.current == nil {
		return
	}
	sm.current.Draw(screen, res)
}
