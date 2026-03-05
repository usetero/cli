package domain

import "testing"

func TestEncodeToolResultsPreservesDistinctValues(t *testing.T) {
	input := []ToolResult{
		{
			ToolUseID: "tool-1",
			Content: map[string]any{
				"rows": []map[string]any{
					{"service_id": "ad"},
				},
			},
		},
		{
			ToolUseID: "tool-2",
			Content: map[string]any{
				"rows": []map[string]any{
					{"service_id": "email"},
				},
			},
		},
	}

	encoded, err := EncodeToolResults(input)
	if err != nil {
		t.Fatalf("EncodeToolResults error: %v", err)
	}

	blocks, err := ParseBlocks(encoded)
	if err != nil {
		t.Fatalf("ParseBlocks error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("len(blocks)=%d, want 2", len(blocks))
	}
	if blocks[0].ToolResult == nil || blocks[1].ToolResult == nil {
		t.Fatalf("expected tool_result blocks, got %#v", blocks)
	}
	if blocks[0].ToolResult.ToolUseID != "tool-1" {
		t.Fatalf("blocks[0].ToolUseID=%q, want tool-1", blocks[0].ToolResult.ToolUseID)
	}
	if blocks[1].ToolResult.ToolUseID != "tool-2" {
		t.Fatalf("blocks[1].ToolUseID=%q, want tool-2", blocks[1].ToolResult.ToolUseID)
	}
}
