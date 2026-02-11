package ui

import "github.com/hajimehoshi/ebiten/v2"

// Scene is the interface for all UI scenes.
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
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

func (sm *SceneManager) Update() error {
	if sm.current == nil {
		return nil
	}
	return sm.current.Update()
}

func (sm *SceneManager) Draw(screen *ebiten.Image) {
	if sm.current == nil {
		return
	}
	sm.current.Draw(screen)
}
