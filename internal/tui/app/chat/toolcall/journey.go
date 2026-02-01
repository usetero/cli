package toolcall

import (
	"fmt"

	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderStartJourney(theme *styles.Theme, input *block.StartJourneyInput, width int) string {
	if input == nil {
		return successText(theme, "Journey started")
	}

	msg := fmt.Sprintf("Started journey: %s", input.Name)
	return successText(theme, msg)
}

func renderEndJourney(theme *styles.Theme, input *block.EndJourneyInput, width int) string {
	if input == nil {
		return successText(theme, "Journey ended")
	}

	return successText(theme, "Journey completed")
}
