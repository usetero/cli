package toolcall

import (
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderSeries(theme *styles.Theme, input *block.ShowSeriesInput, width int) string {
	if input == nil {
		return ""
	}

	// TODO: Implement series visualization
	return mutedText(theme, "Series data displayed", width)
}

func renderTimeSeries(theme *styles.Theme, input *block.ShowTimeSeriesInput, width int) string {
	if input == nil {
		return ""
	}

	// TODO: Implement time series visualization (sparkline or chart)
	return mutedText(theme, "Time series data displayed", width)
}
