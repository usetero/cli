package msgs

import domaintools "github.com/usetero/cli/internal/domain/tools"

// UserSubmittedInput is fired when user input is ready (text from command bar or tool results).
type UserSubmittedInput struct {
	Text        string
	ToolResults []domaintools.Result
}
