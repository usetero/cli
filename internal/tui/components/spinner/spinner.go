// Package spinner provides an animated loading spinner with gradient colors.
// Adapted from Charm's Crush project.
package spinner

import (
	"image/color"
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

const (
	initialChar   = '.'
	labelGap      = " "
	labelGapWidth = 1
)

var availableRunes = []rune("0123456789abcdefABCDEF~!@#$£€%^&*()+=_")
var ellipsisFrames = []string{".", "..", "...", ""}

// Internal ID management ensures tick messages are routed to the correct instance.
var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// TickMsg triggers the next animation frame.
type TickMsg struct{ id int }

// Spinner is an animated loading indicator with gradient colors.
type Spinner struct {
	// Configuration
	width            int
	cyclingCharWidth int
	labelWidth       int
	labelColor       color.Color
	label            []string
	ellipsisFrames   []string

	// Pre-rendered frames
	initialFrames [][]string
	cyclingFrames [][]string

	// Animation state
	id           int
	startTime    time.Time
	birthOffsets []time.Duration
	initialized  atomic.Bool
	step         atomic.Int64
	ellipsisStep atomic.Int64
}

// New creates a new Spinner with the given settings.
func New(opts Settings) *Spinner {
	opts.normalize()

	s := &Spinner{
		id:               nextID(),
		startTime:        time.Now(),
		cyclingCharWidth: opts.Size,
		labelColor:       opts.LabelColor,
	}

	// Check cache first
	key := cacheKey(opts)
	if cached, ok := getCache(key); ok {
		s.width = cached.width
		s.labelWidth = cached.labelWidth
		s.label = make([]string, len(cached.label))
		copy(s.label, cached.label)
		s.ellipsisFrames = make([]string, len(cached.ellipsisFrames))
		copy(s.ellipsisFrames, cached.ellipsisFrames)
		s.initialFrames = cached.initialFrames
		s.cyclingFrames = cached.cyclingFrames
	} else {
		s.buildFrames(opts)

		// Cache the results
		setCache(key, &cache{
			initialFrames:  s.initialFrames,
			cyclingFrames:  s.cyclingFrames,
			width:          s.width,
			labelWidth:     s.labelWidth,
			label:          append([]string{}, s.label...),
			ellipsisFrames: append([]string{}, s.ellipsisFrames...),
		})
	}

	// Random birth offsets for staggered entrance effect
	s.birthOffsets = make([]time.Duration, s.width)
	for i := range s.birthOffsets {
		s.birthOffsets[i] = time.Duration(rand.N(int64(MaxBirthOffset))) * time.Nanosecond
	}

	return s
}

// buildFrames pre-renders all animation frames.
func (s *Spinner) buildFrames(opts Settings) {
	s.labelWidth = lipgloss.Width(opts.Label)

	// Total width
	s.width = opts.Size
	if opts.Label != "" {
		s.width += labelGapWidth + s.labelWidth
	}

	// Pre-render the label
	s.renderLabel(opts.Label)

	// Generate gradient ramp
	var ramp []color.Color
	numFrames := PrerenderedFrames
	if opts.CycleColors {
		ramp = styles.BlendColors(s.width*3, opts.GradColorA, opts.GradColorB, opts.GradColorA, opts.GradColorB)
		numFrames = s.width * 2
	} else {
		ramp = styles.BlendColors(s.width, opts.GradColorA, opts.GradColorB)
	}

	// Pre-render initial characters (dots)
	s.initialFrames = make([][]string, numFrames)
	offset := 0
	for i := range s.initialFrames {
		s.initialFrames[i] = make([]string, s.width+labelGapWidth+s.labelWidth)
		for j := range s.initialFrames[i] {
			if j+offset >= len(ramp) {
				continue
			}

			var c color.Color
			if j <= s.cyclingCharWidth {
				c = ramp[j+offset]
			} else {
				c = s.labelColor
			}

			s.initialFrames[i][j] = lipgloss.NewStyle().
				Foreground(c).
				Render(string(initialChar))
		}
		if opts.CycleColors {
			offset++
		}
	}

	// Pre-render cycling character frames
	s.cyclingFrames = make([][]string, numFrames)
	offset = 0
	for i := range s.cyclingFrames {
		s.cyclingFrames[i] = make([]string, s.width)
		for j := range s.cyclingFrames[i] {
			if j+offset >= len(ramp) {
				continue
			}

			r := availableRunes[rand.IntN(len(availableRunes))]
			s.cyclingFrames[i][j] = lipgloss.NewStyle().
				Foreground(ramp[j+offset]).
				Render(string(r))
		}
		if opts.CycleColors {
			offset++
		}
	}
}

// renderLabel pre-renders the label and ellipsis frames.
func (s *Spinner) renderLabel(label string) {
	if s.labelWidth > 0 {
		labelRunes := []rune(label)
		s.label = make([]string, len(labelRunes))
		for i, r := range labelRunes {
			s.label[i] = lipgloss.NewStyle().
				Foreground(s.labelColor).
				Render(string(r))
		}

		s.ellipsisFrames = make([]string, len(ellipsisFrames))
		for i, frame := range ellipsisFrames {
			s.ellipsisFrames[i] = lipgloss.NewStyle().
				Foreground(s.labelColor).
				Render(frame)
		}
	}
}

// SetLabel updates the label text.
func (s *Spinner) SetLabel(label string) {
	s.labelWidth = lipgloss.Width(label)
	s.width = s.cyclingCharWidth
	if label != "" {
		s.width += labelGapWidth + s.labelWidth
	}
	s.renderLabel(label)
}

// Width returns the total width of the spinner.
func (s *Spinner) Width() int {
	w := s.width
	if s.labelWidth > 0 {
		w += labelGapWidth + s.labelWidth
		// Add widest ellipsis frame
		for _, f := range ellipsisFrames {
			if fw := lipgloss.Width(f); fw > 3 {
				w += fw
			}
		}
		w += 3 // "..."
	}
	return w
}

// Init starts the animation.
func (s *Spinner) Init() tea.Cmd {
	return s.tick()
}

// Update processes animation ticks.
func (s *Spinner) Update(msg tea.Msg) (*Spinner, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.id != s.id {
			return s, nil
		}

		step := s.step.Add(1)
		if int(step) >= len(s.cyclingFrames) {
			s.step.Store(0)
		}

		if s.initialized.Load() && s.labelWidth > 0 {
			ellipsisStep := s.ellipsisStep.Add(1)
			if int(ellipsisStep) >= EllipsisAnimSpeed*len(ellipsisFrames) {
				s.ellipsisStep.Store(0)
			}
		} else if !s.initialized.Load() && time.Since(s.startTime) >= time.Duration(MaxBirthOffset) {
			s.initialized.Store(true)
		}

		return s, s.tick()
	}

	return s, nil
}

// View renders the current animation frame.
func (s *Spinner) View() string {
	var b strings.Builder
	step := int(s.step.Load())

	for i := range s.width {
		switch {
		case !s.initialized.Load() && i < len(s.birthOffsets) && time.Since(s.startTime) < s.birthOffsets[i]:
			// Birth offset not reached: render initial character
			if step < len(s.initialFrames) && i < len(s.initialFrames[step]) {
				b.WriteString(s.initialFrames[step][i])
			}
		case i < s.cyclingCharWidth:
			// Render cycling character
			if step < len(s.cyclingFrames) && i < len(s.cyclingFrames[step]) {
				b.WriteString(s.cyclingFrames[step][i])
			}
		case i == s.cyclingCharWidth:
			// Render label gap
			b.WriteString(labelGap)
		case i > s.cyclingCharWidth:
			// Render label character
			labelIdx := i - s.cyclingCharWidth - labelGapWidth
			if labelIdx >= 0 && labelIdx < len(s.label) {
				b.WriteString(s.label[labelIdx])
			}
		}
	}

	// Render ellipsis after label
	if s.initialized.Load() && s.labelWidth > 0 && len(s.ellipsisFrames) > 0 {
		ellipsisStep := int(s.ellipsisStep.Load())
		frameIdx := ellipsisStep / EllipsisAnimSpeed
		if frameIdx >= 0 && frameIdx < len(s.ellipsisFrames) {
			b.WriteString(s.ellipsisFrames[frameIdx])
		}
	}

	return b.String()
}

// tick returns a command that sends the next tick.
func (s *Spinner) tick() tea.Cmd {
	return tea.Tick(time.Second/DefaultFPS, func(t time.Time) tea.Msg {
		return TickMsg{id: s.id}
	})
}
