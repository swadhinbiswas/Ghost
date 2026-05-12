package model

import (
	"github.com/swadhinbiswas/ghost/internal/ui/multiplexer"
)

// mpAdapter adapts the UI multiplexer to the agent.PaneController interface.
type mpAdapter struct {
	m *multiplexer.Multiplexer
}

func newMPAdapter(m *multiplexer.Multiplexer) *mpAdapter {
	return &mpAdapter{m: m}
}

func (a *mpAdapter) CreatePane(title, content string) string {
	return a.m.CreatePane(title, content)
}

func (a *mpAdapter) UpdateContent(id, content string) {
	a.m.UpdateContent(id, content)
}

func (a *mpAdapter) AppendContent(id, content string) {
	a.m.AppendContent(id, content)
}

func (a *mpAdapter) ClosePane(id string) {
	pane := a.m.GetPane(id)
	if pane != nil {
		pane.Visible = false
	}
}

func (a *mpAdapter) PaneCount() int {
	return a.m.PaneCount()
}
