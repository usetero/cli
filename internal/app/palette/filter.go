package palette

import "github.com/sahilm/fuzzy"

// pushLevel saves the current level and drills into children.
// title is the name of the command being drilled into (shown as header).
func (m *Model) pushLevel(title string, children []Command) {
	m.stack = append(m.stack, level{
		commands: m.commands,
		title:    title,
	})
	m.commands = children
	m.selected = 0
	m.input.SetValue("")
	m.filter()
}

// popLevel returns to the previous level.
func (m *Model) popLevel() {
	prev := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	m.commands = prev.commands
	m.selected = 0
	m.input.SetValue("")
	m.filter()
}

// filter updates the match list based on current input.
func (m *Model) filter() {
	query := m.input.Value()

	if query == "" {
		// Show all commands.
		m.matches = make([]match, len(m.commands))
		for i, cmd := range m.commands {
			m.matches[i] = match{command: cmd}
		}
	} else {
		// Fuzzy match.
		source := commandSource(m.commands)
		results := fuzzy.FindFrom(query, source)
		m.matches = make([]match, len(results))
		for i, r := range results {
			m.matches[i] = match{
				command:      m.commands[r.Index],
				matchIndexes: r.MatchedIndexes,
			}
		}
	}

	// Clamp selection.
	if m.selected >= len(m.matches) {
		m.selected = max(0, len(m.matches)-1)
	}
}

// commandSource implements fuzzy.Source for a slice of commands.
type commandSource []Command

func (s commandSource) String(i int) string { return s[i].Name }
func (s commandSource) Len() int            { return len(s) }
