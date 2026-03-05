package tools

// Name identifies a chat tool in both local runtime and chat API payloads.
type Name string

const (
	QueryToolName         Name = "query"
	ShowToolName          Name = "show"
	EnableServiceToolName Name = "enable_service"
	ApprovePolicyToolName Name = "approve_policy"
)

// Definition is the tool contract sent to the chat API.
type Definition struct {
	Name        Name
	Description string
	InputSchema map[string]any
}
