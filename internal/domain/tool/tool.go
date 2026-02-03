package tool

// Name identifies a known tool.
type Name string

const (
	AddContext     Name = "add_context"
	RemoveContext  Name = "remove_context"
	Query          Name = "query"
	ShowMetric     Name = "show_metric"
	ShowSeries     Name = "show_series"
	ShowTimeSeries Name = "show_time_series"
	ShowTable      Name = "show_table"
	StartJourney   Name = "start_journey"
	EndJourney     Name = "end_journey"
	ApprovePolicy  Name = "approve_policy"
	DismissPolicy  Name = "dismiss_policy"
)

// Executor identifies who is responsible for executing a tool call.
type Executor string

const (
	// ExecutorServer means the server executes this tool and streams the result.
	ExecutorServer Executor = "server"

	// ExecutorClient means the client must execute this tool and send the result.
	ExecutorClient Executor = "client"
)

// Executor returns who is responsible for executing this tool.
func (n Name) Executor() Executor {
	switch n {
	case AddContext, RemoveContext, Query, StartJourney, EndJourney, ApprovePolicy, DismissPolicy:
		return ExecutorServer
	case ShowMetric, ShowSeries, ShowTable, ShowTimeSeries:
		return ExecutorClient
	}
	return ExecutorClient
}

// Use represents the AI calling a tool.
// Exactly one of the typed input fields is populated, determined by Name.
type Use struct {
	ID             string               `json:"id"`
	Name           Name                 `json:"name"`
	ExecutedBy     Executor             `json:"executed_by,omitempty"`
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
func (t Use) IsComplete() bool {
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

// Result is the outcome of a tool call.
type Result struct {
	ToolUseID     string               `json:"tool_use_id"`
	IsError       bool                 `json:"is_error,omitempty"`
	Error         string               `json:"error,omitempty"`
	AddContext    *AddContextResult    `json:"add_context,omitempty"`
	RemoveContext *RemoveContextResult `json:"remove_context,omitempty"`
	Query         *QueryResult         `json:"query,omitempty"`
	ApprovePolicy *ApprovePolicyResult `json:"approve_policy,omitempty"`
	DismissPolicy *DismissPolicyResult `json:"dismiss_policy,omitempty"`
}
