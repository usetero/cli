package toolcall

import (
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderTable(theme *styles.Theme, input *block.ShowTableInput, width int) string {
	if input == nil {
		return ""
	}

	// TODO: Implement table visualization
	return mutedText(theme, "Table data displayed", width)
}
