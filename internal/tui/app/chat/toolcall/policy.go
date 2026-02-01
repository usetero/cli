package toolcall

import (
	"fmt"

	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
)

func renderApprovePolicy(theme *styles.Theme, input *block.ApprovePolicyInput, result *block.ApprovePolicyResult, width int) string {
	if input == nil {
		return successText(theme, "Policy approved")
	}

	msg := fmt.Sprintf("Approved policy: %s", input.PolicyID)
	return successText(theme, msg)
}

func renderDismissPolicy(theme *styles.Theme, input *block.DismissPolicyInput, result *block.DismissPolicyResult, width int) string {
	if input == nil {
		return successText(theme, "Policy dismissed")
	}

	msg := fmt.Sprintf("Dismissed policy: %s", input.PolicyID)
	return successText(theme, msg)
}
