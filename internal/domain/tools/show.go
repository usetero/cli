package tools

// EntityType identifies the kind of entity to show.
type EntityType string

const (
	EntityPolicy EntityType = "policy"
)

// ShowInput is the input schema for the show tool.
type ShowInput struct {
	Entity EntityType `json:"entity"`
	ID     string     `json:"id,omitempty"`
	SQL    string     `json:"sql,omitempty"`
	Title  string     `json:"title,omitempty"`
}

// ShowResult is the typed output of a show tool execution.
type ShowResult struct {
	Entity      EntityType     `json:"entity"`
	ID          string         `json:"id"`
	IDShort     string         `json:"id_short"`
	Title       string         `json:"title,omitempty"`
	CardSummary string         `json:"card_summary"`
	Data        map[string]any `json:"data"`
}

// ToMap serializes the result for the GraphQL API.
// The card_summary tells the AI what the user sees on screen so it can
// reference the card conversationally without repeating its contents.
func (r ShowResult) ToMap() map[string]any {
	return map[string]any{
		"displayed":    true,
		"entity":       string(r.Entity),
		"id":           r.ID,
		"id_short":     r.IDShort,
		"title":        r.Title,
		"card_summary": r.CardSummary,
		"data":         r.Data,
	}
}
