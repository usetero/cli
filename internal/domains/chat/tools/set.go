package tools

import (
	"context"
	"encoding/json"
)

// Toolset is a static, typed tool collection used by chat runtime.
type Toolset struct {
	Query         *QueryTool
	Show          *ShowTool
	EnableService *EnableServiceTool
}

func (s Toolset) Definitions() []Definition {
	out := make([]Definition, 0, 3)
	if s.Query != nil {
		out = append(out, s.Query.Definition())
	}
	if s.Show != nil {
		out = append(out, s.Show.Definition())
	}
	if s.EnableService != nil {
		out = append(out, s.EnableService.Definition())
	}
	return out
}

func (s Toolset) Run(ctx context.Context, name Name, input json.RawMessage) (json.RawMessage, error, bool) {
	switch name {
	case QueryToolName:
		if s.Query == nil {
			return nil, nil, false
		}
		out, err := s.Query.Run(ctx, input)
		return out, err, true
	case ShowToolName:
		if s.Show == nil {
			return nil, nil, false
		}
		out, err := s.Show.Run(ctx, input)
		return out, err, true
	case EnableServiceToolName:
		if s.EnableService == nil {
			return nil, nil, false
		}
		out, err := s.EnableService.Run(ctx, input)
		return out, err, true
	}
	return nil, nil, false
}
