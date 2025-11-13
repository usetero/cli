package sidebar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/logo"
	"github.com/usetero/cli/internal/tui/styles"
)

const diag = `╱`

// Component represents the chat sidebar
type Component struct {
	width  int
	height int
	logger log.Logger

	// Context information
	orgName string

	// Catalog stats (TODO: these will come from the control plane)
	servicesCount int
	logsRate      string // e.g. "1.54m/hr"

	// Quality stats
	wastePercent int
	wasteTrend   string // e.g. "↑2%" or "↓4%"
	savedAmount  string // e.g. "$847k/yr"

	// Contracts stats
	renewalDays   int    // Days until renewal (negative means days remaining)
	renewalAmount string // e.g. "$2.3m"

	// User info
	userName  string
	userEmail string
}

// New creates a new sidebar component
func New(logger log.Logger) Component {
	return Component{
		logger: logger,
		// TODO: These will be passed in from the chat page / control plane
		orgName:       "Acme Corp",
		servicesCount: 2,
		logsRate:      "1.54m/hr",
		wastePercent:  23,
		wasteTrend:    "↑2%",
		savedAmount:   "$847k/yr",
		renewalDays:   -23,
		renewalAmount: "$2.3m",
		userName:      "Ben Johnson",
		userEmail:     "ben@acme.com",
	}
}

// SetSize sets the dimensions for the sidebar
func (c *Component) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// renderSection creates a section header with a text label followed by a line
// Example: "Context ─────────────────"
func (c *Component) renderSection(text string, theme *styles.Theme) string {
	char := "─"
	length := lipgloss.Width(text) + 1
	remainingWidth := c.width - length
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	lineStyle := lipgloss.NewStyle().Foreground(theme.Border)
	return text + " " + lineStyle.Render(strings.Repeat(char, remainingWidth))
}

// Render renders the sidebar
func (c *Component) Render() string {
	if c.width == 0 || c.height == 0 {
		return ""
	}

	theme := styles.CurrentTheme()

	// Container style for the entire sidebar
	style := lipgloss.NewStyle().
		Width(c.width).
		Height(c.height)

	// Create diagonal lines that span the full sidebar width
	fieldStyle := lipgloss.NewStyle().Foreground(theme.Field)
	divider := fieldStyle.Render(strings.Repeat(diag, c.width))

	// Render ASCII logo
	logoView := logo.Render(logo.Opts{
		TitleColorA: theme.Primary,
		TitleColorB: theme.Secondary,
	})

	// Split logo into lines so we can add version to the last line
	logoLines := strings.Split(logoView, "\n")

	// Version text (right-aligned on same line as last logo line)
	versionText := lipgloss.NewStyle().
		Foreground(theme.Field).
		Render("v0.0.1")

	// Add version to the last line of the logo
	if len(logoLines) > 0 {
		lastLine := logoLines[len(logoLines)-1]
		lastLineWidth := lipgloss.Width(lastLine)

		// Calculate spacing to right-align version
		spacingWidth := c.width - lastLineWidth - lipgloss.Width(versionText)
		if spacingWidth < 0 {
			spacingWidth = 0
		}

		logoLines[len(logoLines)-1] = lastLine + strings.Repeat(" ", spacingWidth) + versionText
	}

	// Rejoin logo with version
	logoWithVersion := strings.Join(logoLines, "\n")

	// Section headers (no header for org name - it's self-explanatory)
	navigationHeader := c.renderSection("Navigation", theme)
	catalogHeader := c.renderSection("Catalog", theme)
	contractsHeader := c.renderSection("Contracts", theme)

	// Org/Account section (no header, just the name)
	orgStyle := lipgloss.NewStyle().Foreground(theme.Text)
	orgName := orgStyle.Render(c.orgName)
	// TODO: Add accountName and workspace if > 1

	// User info (right under org)
	userNameStyle := lipgloss.NewStyle().Foreground(theme.Text)
	userEmailStyle := lipgloss.NewStyle().Foreground(theme.Field)

	// Define key bindings for navigation
	chatKey := key.NewBinding(
		key.WithKeys("alt+1"),
		key.WithHelp("⌥1", "Chat"),
	)
	servicesKey := key.NewBinding(
		key.WithKeys("alt+2"),
		key.WithHelp("⌥2", "Services"),
	)
	logsKey := key.NewBinding(
		key.WithKeys("alt+3"),
		key.WithHelp("⌥3", "Logs"),
	)
	wasteKey := key.NewBinding(
		key.WithKeys("alt+4"),
		key.WithHelp("⌥4", "Waste"),
	)
	savedKey := key.NewBinding(
		key.WithKeys("alt+5"),
		key.WithHelp("⌥5", "Saved"),
	)
	renewalKey := key.NewBinding(
		key.WithKeys("alt+6"),
		key.WithHelp("⌥6", "DD Renewal"),
	)

	// Navigation section - Chat only
	chatItem := NewNavItem("Chat", "", nil, true, false, chatKey) // Active, no indicator

	// Catalog section - Services, Logs, Waste
	servicesItem := NewNavItem("Services", fmt.Sprintf("%d", c.servicesCount), nil, false, false, servicesKey)
	logsItem := NewNavItem("Logs", c.logsRate, nil, false, false, logsKey)
	// Waste is red because 23% is over the typical 10% goal (in reality, this would come from control plane)
	// The 'w' suffix indicates week-over-week change
	wasteItem := NewNavItem("Waste", fmt.Sprintf("%d%% %sw", c.wastePercent, c.wasteTrend), theme.Error, false, true, wasteKey) // Show indicator for demo

	// Contracts section - Saved, DD Renewal
	// Saved is green because it's a positive outcome (money saved)
	savedItem := NewNavItem("Saved", c.savedAmount, theme.Success, false, false, savedKey)
	// Renewal date is red because -23d means renewing 23 days early (bad)
	renewalItem := NewNavItem("DD Renewal", fmt.Sprintf("%dd, %s", c.renewalDays, c.renewalAmount), theme.Error, false, false, renewalKey)

	// All content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		divider,
		divider,
		"",
		logoWithVersion,
		"",
		orgName,
		userNameStyle.Render(c.userName),
		userEmailStyle.Render(c.userEmail),
		"",
		navigationHeader,
		"",
		chatItem.Render(c.width, theme),
		"",
		catalogHeader,
		"",
		servicesItem.Render(c.width, theme),
		logsItem.Render(c.width, theme),
		wasteItem.Render(c.width, theme),
		"",
		contractsHeader,
		"",
		savedItem.Render(c.width, theme),
		renewalItem.Render(c.width, theme),
	)

	return style.Render(content)
}
