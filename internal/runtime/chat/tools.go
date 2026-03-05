package chat

import (
	"context"
	"fmt"
	"strings"

	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
)

func (r *Runtime) toolDefinitions() []infrachat.Tool {
	defs := r.tools.Definitions()
	out := make([]infrachat.Tool, 0, len(defs))
	for i := range defs {
		def := defs[i]
		if def.Name == "" {
			continue
		}
		out = append(out, infrachat.Tool{
			Name:        toWireToolName(def.Name),
			Description: def.Description,
			InputSchema: def.InputSchema,
		})
	}
	return out
}

func (r *Runtime) executeToolUses(ctx context.Context, toolUses []infrachat.ToolUse) ([]infrachat.ToolResult, string) {
	results := make([]infrachat.ToolResult, 0, len(toolUses))
	summaryParts := make([]string, 0, len(toolUses))
	for i := range toolUses {
		toolUse := toolUses[i]
		result := infrachat.ToolResult{ToolUseID: toolUse.ID}

		output, err, ok := r.tools.Run(ctx, toDomainToolName(toolUse.Name), toolUse.Input)
		if !ok {
			result.IsError = true
			result.Error = fmt.Sprintf("unknown tool: %s", toolUse.Name)
			summaryParts = append(summaryParts, toolUse.Name+": error")
			results = append(results, result)
			continue
		}
		if err != nil {
			result.IsError = true
			result.Error = err.Error()
			summaryParts = append(summaryParts, toolUse.Name+": error")
		} else {
			result.Content = normalizeToolOutput(output)
			summaryParts = append(summaryParts, toolUse.Name+": ok")
		}
		results = append(results, result)
	}

	if len(summaryParts) == 0 {
		return results, "tool run complete"
	}
	return results, "tool run: " + strings.Join(summaryParts, ", ")
}
