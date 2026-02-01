package toolcall

import (
	"fmt"

	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderAddContext(theme *styles.Theme, input *block.AddContextInput, result *block.AddContextResult, width int) string {
	if input == nil {
		return successText(theme, "Context added")
	}

	msg := fmt.Sprintf("Added %s: %s", input.EntityType, input.EntityID)
	return successText(theme, msg)
}

func renderRemoveContext(theme *styles.Theme, input *block.RemoveContextInput, result *block.RemoveContextResult, width int) string {
	if input == nil {
		return successText(theme, "Context removed")
	}

	msg := fmt.Sprintf("Removed %s: %s", input.EntityType, input.EntityID)
	return successText(theme, msg)
}
