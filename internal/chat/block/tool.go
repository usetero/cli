package block

// ToolName identifies a known tool.
type ToolName string

const (
	ToolAddContext     ToolName = "add_context"
	ToolRemoveContext  ToolName = "remove_context"
	ToolQuery          ToolName = "query"
	ToolShowMetric     ToolName = "show_metric"
	ToolShowSeries     ToolName = "show_series"
	ToolShowTimeSeries ToolName = "show_time_series"
	ToolShowTable      ToolName = "show_table"
	ToolStartJourney   ToolName = "start_journey"
	ToolEndJourney     ToolName = "end_journey"
	ToolApprovePolicy  ToolName = "approve_policy"
	ToolDismissPolicy  ToolName = "dismiss_policy"
)

// ToolExecutor identifies who is responsible for executing a tool call.
type ToolExecutor string

const (
	// ExecutorServer means the server executes this tool and streams the result.
	ExecutorServer ToolExecutor = "server"

	// ExecutorClient means the client must execute this tool and send the result.
	ExecutorClient ToolExecutor = "client"
)

// Executor returns who is responsible for executing this tool.
func (n ToolName) Executor() ToolExecutor {
	switch n {
	case ToolAddContext, ToolRemoveContext, ToolQuery, ToolStartJourney, ToolEndJourney, ToolApprovePolicy, ToolDismissPolicy:
		return ExecutorServer
	case ToolShowMetric, ToolShowSeries, ToolShowTable, ToolShowTimeSeries:
		return ExecutorClient
	}
	return ExecutorClient
}

// ToolUse represents the AI calling a tool.
// Exactly one of the typed input fields is populated, determined by Name.
// During streaming, RawInput accumulates partial JSON before being parsed.
type ToolUse struct {
	ID             string               `json:"id"`
	Name           ToolName             `json:"name"`
	ExecutedBy     ToolExecutor         `json:"executed_by,omitempty"`
	RawInput       string               `json:"raw_input,omitempty"` // Streaming only - accumulated JSON
	AddContext     *AddContextInput     `json:"add_context,omitempty"`
	RemoveContext  *RemoveContextInput  `json:"remove_context,omitempty"`
	Query          *QueryInput          `json:"query,omitempty"`
	ShowMetric     *ShowMetricInput     `json:"show_metric,omitempty"`
	ShowSeries     *ShowSeriesInput     `json:"show_series,omitempty"`
	ShowTimeSeries *ShowTimeSeriesInput `json:"show_time_series,omitempty"`
	ShowTable      *ShowTableInput      `json:"show_table,omitempty"`
	StartJourney   *StartJourneyInput   `json:"start_journey,omitempty"`
	EndJourney     *EndJourneyInput     `json:"end_journey,omitempty"`
	ApprovePolicy  *ApprovePolicyInput  `json:"approve_policy,omitempty"`
	DismissPolicy  *DismissPolicyInput  `json:"dismiss_policy,omitempty"`
}

// IsComplete returns true if the tool use has its typed input populated.
func (t ToolUse) IsComplete() bool {
	return t.AddContext != nil ||
		t.RemoveContext != nil ||
		t.Query != nil ||
		t.ShowMetric != nil ||
		t.ShowSeries != nil ||
		t.ShowTimeSeries != nil ||
		t.ShowTable != nil ||
		t.StartJourney != nil ||
		t.EndJourney != nil ||
		t.ApprovePolicy != nil ||
		t.DismissPolicy != nil
}

// ToolResult is the outcome of a tool call.
type ToolResult struct {
	ToolUseID     string               `json:"tool_use_id"`
	IsError       bool                 `json:"is_error,omitempty"`
	Error         string               `json:"error,omitempty"`
	AddContext    *AddContextResult    `json:"add_context,omitempty"`
	RemoveContext *RemoveContextResult `json:"remove_context,omitempty"`
	Query         *QueryResult         `json:"query,omitempty"`
	ApprovePolicy *ApprovePolicyResult `json:"approve_policy,omitempty"`
	DismissPolicy *DismissPolicyResult `json:"dismiss_policy,omitempty"`
}
