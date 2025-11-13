package datadog

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/tui/styles"
)

// RenderHeader renders the consistent Datadog onboarding header
func RenderHeader() string {
<<<<<<< HEAD
	theme := styles.CurrentTheme()

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	bodyStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	mutedStyle := lipgloss.NewStyle().
		Foreground(theme.TextMuted)

=======
	common := styles.Common()
	theme := styles.CurrentTheme()

>>>>>>> 17e8dd9 (chore: initial commit)
	linkStyle := lipgloss.NewStyle().
		Foreground(theme.TextSubtle).
		Underline(true)

	checkStyle := lipgloss.NewStyle().
		Foreground(theme.Success)

	crossStyle := lipgloss.NewStyle().
		Foreground(theme.Error)

<<<<<<< HEAD
	title := titleStyle.Render("Connect your Datadog account")

	intro := bodyStyle.Render("Tero builds a complete understanding of your observability data—what") + "\n" +
		bodyStyle.Render("it means, what's valuable, and what's waste—so you can improve data") + "\n" +
		bodyStyle.Render("quality, reduce cost, and give your engineers better data.")

	whatTeroDoes := bodyStyle.Render("What Tero does:")
	doesLine1 := checkStyle.Render(" ✓") + " " + mutedStyle.Render("Analyzes your logs to understand patterns and meaning")
	doesLine2 := checkStyle.Render(" ✓") + " " + mutedStyle.Render("Builds a semantic catalog of your telemetry")
	doesLine3 := checkStyle.Render(" ✓") + " " + mutedStyle.Render("Identifies quality issues and waste patterns")

	whatTeroNever := bodyStyle.Render("What Tero NEVER does:")
	neverLine1 := crossStyle.Render(" ✗") + " " + mutedStyle.Render("Stores your log data (we analyze, don't persist)")
	neverLine2 := crossStyle.Render(" ✗") + " " + mutedStyle.Render("Changes anything without your explicit approval")
	neverLine3 := crossStyle.Render(" ✗") + " " + mutedStyle.Render("Requires infrastructure changes or deployment")

	safety := mutedStyle.Render("Read-only access. Fully reversible.")

	carrotStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	carrot := carrotStyle.Render("Most teams find 40% waste in 5 minutes.")

	learnMore := mutedStyle.Render("Learn more: ") + linkStyle.Render("usetero.com/docs/integrations/datadog") + mutedStyle.Render(" (press o to open)")
=======
	title := common.Title.Render("Connect your Datadog account")

	intro := common.Body.Render("Tero builds a complete understanding of your observability data—what") + "\n" +
		common.Body.Render("it means, what's valuable, and what's waste—so you can improve data") + "\n" +
		common.Body.Render("quality, reduce cost, and give your engineers better data.")

	whatTeroDoes := common.Body.Render("What Tero does:")
	doesLine1 := checkStyle.Render(" ✓") + " " + common.Help.Render("Analyzes your logs to understand patterns and meaning")
	doesLine2 := checkStyle.Render(" ✓") + " " + common.Help.Render("Builds a semantic catalog of your telemetry")
	doesLine3 := checkStyle.Render(" ✓") + " " + common.Help.Render("Identifies quality issues and waste patterns")

	whatTeroNever := common.Body.Render("What Tero NEVER does:")
	neverLine1 := crossStyle.Render(" ✗") + " " + common.Help.Render("Stores your log data (we analyze, don't persist)")
	neverLine2 := crossStyle.Render(" ✗") + " " + common.Help.Render("Changes anything without your explicit approval")
	neverLine3 := crossStyle.Render(" ✗") + " " + common.Help.Render("Requires infrastructure changes or deployment")

	safety := common.Help.Render("Read-only access. Fully reversible.")
	carrot := common.Title.Render("Most teams find 40% waste in 5 minutes.")
	learnMore := common.Help.Render("Learn more: ") + linkStyle.Render("usetero.com/docs/integrations/datadog") + common.Help.Render(" (press o to open)")
>>>>>>> 17e8dd9 (chore: initial commit)

	divider := lipgloss.NewStyle().
		Foreground(theme.Border).
		Render("────────────────────────────────────────────────────────")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		intro,
		"",
		whatTeroDoes,
		doesLine1,
		doesLine2,
		doesLine3,
		"",
		whatTeroNever,
		neverLine1,
		neverLine2,
		neverLine3,
		"",
		safety,
		"",
		carrot,
		"",
		learnMore,
		"",
		divider,
		"",
	)
}
