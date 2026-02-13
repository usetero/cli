package tools

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain/tools"
)

// StartPolicyApprovalTool triggers the interactive policy approval wizard.
// Unlike other tools, this doesn't execute immediately - the UI block
// emits a message to open the wizard, then waits for completion.
type StartPolicyApprovalTool struct{}

// NewStartPolicyApprovalTool creates a new start_policy_approval tool.
func NewStartPolicyApprovalTool() *StartPolicyApprovalTool {
	return &StartPolicyApprovalTool{}
}

// Name returns the tool name.
func (t *StartPolicyApprovalTool) Name() string {
	return "start_policy_approval"
}

// Definition returns the tool definition for the chat API.
func (t *StartPolicyApprovalTool) Definition() chat.Tool {
	return chat.Tool{
		Name: t.Name(),
		Description: `Start an interactive policy approval flow.

This opens a wizard where the user can:
- Review all pending log event policies
- Select which policies to approve
- Confirm the approval with a summary of impact

Use this when the user wants to:
- Review and approve pending policies
- Bulk approve multiple policies
- See what policies are available for approval

IMPORTANT: Call this tool immediately without any preceding text. Do not explain what you're about to do - just call the tool directly.

The tool returns after the user completes or cancels the wizard.`,
		InputSchema: chat.NewObjectSchema(map[string]chat.Property{}, nil),
	}
}

// Execute is a no-op for this tool. The actual execution happens in the UI block
// which emits StartPolicyApprovalMsg to trigger the wizard.
func (t *StartPolicyApprovalTool) Execute(input json.RawMessage) (tools.StartPolicyApprovalResult, error) {
	// This is never called directly - the UI block handles everything
	return tools.StartPolicyApprovalResult{}, nil
}
