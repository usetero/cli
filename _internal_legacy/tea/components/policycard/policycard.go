// Package policycard renders a policy as a self-contained card component.
// Designed for embedding in chat — the card stands on its own with identity,
// rationale, recommendation, impact, and evidence from actual log data.
package policycard

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

// Model renders a single policy card.
type Model struct {
	theme  styles.Theme
	width  int
	policy *domain.Policy

	// Cached on SetPolicy — derived once, used in View.
	impact   *domain.PolicyImpact
	evidence domain.Evidence
}

// New creates a policy card. Call SetPolicy and SetWidth before rendering.
func New(theme styles.Theme) *Model {
	return &Model{theme: theme}
}

// SetPolicy sets the policy to render and caches derived data.
func (m *Model) SetPolicy(p *domain.Policy) {
	m.policy = p
	if p != nil {
		m.impact = p.Impact()
		m.evidence = domain.BuildEvidence(p)
	} else {
		m.impact = nil
		m.evidence = nil
	}
}

// SetWidth sets the available rendering width.
func (m *Model) SetWidth(w int) {
	m.width = w
}

// Height returns the rendered height in lines.
func (m *Model) Height() int {
	if m.policy == nil {
		return 0
	}
	return strings.Count(m.View(), "\n") + 1
}

// Update handles messages. No interactive state yet.
func (m *Model) Update(_ tea.Msg) tea.Cmd {
	return nil
}

// View renders the complete policy card.
func (m *Model) View() string {
	if m.policy == nil {
		return ""
	}
	if m.width < 20 {
		m.width = 20
	}

	var sections []string

	sections = append(sections, m.viewHeader())

	if s := m.viewRationale(); s != "" {
		sections = append(sections, s)
	}

	sections = append(sections, m.viewRecommendation())

	if s := m.viewImpact(); s != "" {
		sections = append(sections, s)
	}

	if s := m.viewEvidence(); s != "" {
		sections = append(sections, s)
	}

	return strings.Join(sections, "\n\n")
}
