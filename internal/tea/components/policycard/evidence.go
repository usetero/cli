package policycard

import "github.com/usetero/cli/internal/domain"

// viewEvidence renders the evidence section for the policy card.
// Each evidence type owns its complete section rendering — divider,
// optional log body, and content.
func (m *Model) viewEvidence() string {
	if m.evidence == nil {
		return ""
	}

	switch ev := m.evidence.(type) {
	case *domain.ConstantVariesEvidence:
		return m.viewConstantVaries(ev)
	case *domain.HighlightedExampleEvidence:
		return m.viewHighlightedExample(ev)
	case *domain.FieldListEvidence:
		return m.viewFieldList(ev)
	default:
		return ""
	}
}

// viewLogBody renders the first example's log body if available.
// Used by example-based evidence types (constant-varies, highlighted).
func (m *Model) viewLogBody() string {
	if len(m.policy.Examples) == 0 || m.policy.Examples[0].Body == "" {
		return ""
	}
	return m.policy.Examples[0].Body
}
