package api_test

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tool"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestMessageService_CreateMessage(t *testing.T) {
	t.Parallel()

	t.Run("sends text block with correct content", func(t *testing.T) {
		t.Parallel()

		var captured gen.CreateMessageInput
		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				captured = input
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleUser,
			Content: []domain.Block{
				domain.NewTextBlock("Hello, world!"),
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("CreateMessage() error = %v", err)
		}

		if captured.Id == nil || *captured.Id != "msg-123" {
			t.Errorf("Id = %v, want %q", captured.Id, "msg-123")
		}
		if captured.ConversationID != "conv-456" {
			t.Errorf("ConversationID = %q, want %q", captured.ConversationID, "conv-456")
		}
		if captured.Role != gen.MessageRoleUser {
			t.Errorf("Role = %v, want %v", captured.Role, gen.MessageRoleUser)
		}
		if len(captured.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(captured.Content))
		}
		if captured.Content[0].Type != gen.ContentBlockTypeText {
			t.Errorf("Content[0].Type = %v, want %v", captured.Content[0].Type, gen.ContentBlockTypeText)
		}
		if captured.Content[0].Text.Content != "Hello, world!" {
			t.Errorf("Content[0].Text.Content = %q, want %q", captured.Content[0].Text.Content, "Hello, world!")
		}
	})

	t.Run("sends assistant message with model and stop reason", func(t *testing.T) {
		t.Parallel()

		var captured gen.CreateMessageInput
		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				captured = input
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleAssistant,
			Model:          "claude-3",
			StopReason:     "end_turn",
			Content: []domain.Block{
				domain.NewTextBlock("Hello!"),
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("CreateMessage() error = %v", err)
		}

		if captured.Role != gen.MessageRoleAssistant {
			t.Errorf("Role = %v, want %v", captured.Role, gen.MessageRoleAssistant)
		}
		if captured.Model == nil || *captured.Model != "claude-3" {
			t.Errorf("Model = %v, want %q", captured.Model, "claude-3")
		}
		if captured.StopReason == nil || *captured.StopReason != gen.MessageStopReasonEndTurn {
			t.Errorf("StopReason = %v, want %v", captured.StopReason, gen.MessageStopReasonEndTurn)
		}
	})

	t.Run("sends tool_use block with typed input", func(t *testing.T) {
		t.Parallel()

		var captured gen.CreateMessageInput
		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				captured = input
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleAssistant,
			Content: []domain.Block{
				{
					Type: domain.BlockTypeToolUse,
					ToolUse: &tool.Use{
						ID:   "tool-1",
						Name: tool.Query,
						Query: &tool.QueryInput{
							SQL: "SELECT * FROM logs",
						},
					},
				},
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("CreateMessage() error = %v", err)
		}

		if len(captured.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(captured.Content))
		}
		block := captured.Content[0]
		if block.Type != gen.ContentBlockTypeToolUse {
			t.Errorf("Type = %v, want %v", block.Type, gen.ContentBlockTypeToolUse)
		}
		if block.ToolUse.Id != "tool-1" {
			t.Errorf("ToolUse.Id = %q, want %q", block.ToolUse.Id, "tool-1")
		}
		if block.ToolUse.Name != "query" {
			t.Errorf("ToolUse.Name = %q, want %q", block.ToolUse.Name, "query")
		}
		sql, ok := block.ToolUse.Input["sql"].(string)
		if !ok || sql != "SELECT * FROM logs" {
			t.Errorf("ToolUse.Input[sql] = %v, want %q", block.ToolUse.Input["sql"], "SELECT * FROM logs")
		}
	})

	t.Run("rejects tool_use with nil typed input", func(t *testing.T) {
		t.Parallel()

		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				t.Error("CreateMessage should not be called")
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleAssistant,
			Content: []domain.Block{
				{
					Type: domain.BlockTypeToolUse,
					ToolUse: &tool.Use{
						ID:    "tool-1",
						Name:  tool.Query,
						Query: nil, // Missing typed input
					},
				},
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err == nil {
			t.Fatal("CreateMessage() expected error for nil typed input, got nil")
		}
	})

	t.Run("rejects unknown tool name", func(t *testing.T) {
		t.Parallel()

		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				t.Error("CreateMessage should not be called")
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleAssistant,
			Content: []domain.Block{
				{
					Type: domain.BlockTypeToolUse,
					ToolUse: &tool.Use{
						ID:   "tool-1",
						Name: tool.Name("unknown_tool"),
					},
				},
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err == nil {
			t.Fatal("CreateMessage() expected error for unknown tool, got nil")
		}
	})

	t.Run("rejects streaming block types", func(t *testing.T) {
		t.Parallel()

		streamingTypes := []domain.BlockType{
			domain.BlockTypeTextDelta,
			domain.BlockTypeThinkingDelta,
			domain.BlockTypeToolInputDelta,
			domain.BlockTypeMessageStart,
			domain.BlockTypeMessageStop,
		}

		for _, blockType := range streamingTypes {
			t.Run(string(blockType), func(t *testing.T) {
				t.Parallel()

				mockClient := &apitest.MockClient{
					CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
						t.Error("CreateMessage should not be called")
						return &gen.CreateMessageResponse{}, nil
					},
				}

				svc := api.NewMessageService(mockClient, logtest.New(t))

				msg := &domain.Message{
					ID:             "msg-123",
					ConversationID: "conv-456",
					Role:           domain.RoleAssistant,
					Content: []domain.Block{
						{Type: blockType},
					},
				}

				err := svc.CreateMessage(context.Background(), msg)
				if err == nil {
					t.Fatalf("CreateMessage() expected error for streaming block type %s, got nil", blockType)
				}
			})
		}
	})

	t.Run("sends tool_result with typed content", func(t *testing.T) {
		t.Parallel()

		var captured gen.CreateMessageInput
		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				captured = input
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleUser,
			Content: []domain.Block{
				{
					Type: domain.BlockTypeToolResult,
					ToolResult: &tool.Result{
						ToolUseID: "tool-1",
						Query: &tool.QueryResult{
							Columns: []string{"id", "name"},
							Rows:    [][]any{{"1", "foo"}},
						},
					},
				},
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("CreateMessage() error = %v", err)
		}

		if len(captured.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(captured.Content))
		}
		block := captured.Content[0]
		if block.Type != gen.ContentBlockTypeToolResult {
			t.Errorf("Type = %v, want %v", block.Type, gen.ContentBlockTypeToolResult)
		}
		if block.ToolResult.ToolUseId != "tool-1" {
			t.Errorf("ToolResult.ToolUseId = %q, want %q", block.ToolResult.ToolUseId, "tool-1")
		}
		if block.ToolResult.Content == nil {
			t.Fatal("ToolResult.Content is nil")
		}
		cols, ok := (*block.ToolResult.Content)["columns"].([]any)
		if !ok || len(cols) != 2 {
			t.Errorf("ToolResult.Content[columns] = %v, want [id, name]", (*block.ToolResult.Content)["columns"])
		}
	})

	t.Run("sends tool_result error without content", func(t *testing.T) {
		t.Parallel()

		var captured gen.CreateMessageInput
		mockClient := &apitest.MockClient{
			CreateMessageFunc: func(ctx context.Context, input gen.CreateMessageInput) (*gen.CreateMessageResponse, error) {
				captured = input
				return &gen.CreateMessageResponse{}, nil
			},
		}

		svc := api.NewMessageService(mockClient, logtest.New(t))

		msg := &domain.Message{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           domain.RoleUser,
			Content: []domain.Block{
				{
					Type: domain.BlockTypeToolResult,
					ToolResult: &tool.Result{
						ToolUseID: "tool-1",
						IsError:   true,
						Error:     "something went wrong",
					},
				},
			},
		}

		err := svc.CreateMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("CreateMessage() error = %v", err)
		}

		if len(captured.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(captured.Content))
		}
		block := captured.Content[0]
		if !block.ToolResult.IsError {
			t.Error("ToolResult.IsError = false, want true")
		}
		if block.ToolResult.Error == nil || *block.ToolResult.Error != "something went wrong" {
			t.Errorf("ToolResult.Error = %v, want %q", block.ToolResult.Error, "something went wrong")
		}
	})
}
