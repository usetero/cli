package chat

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/internal/domain"
)

func TestToWireRequest_StripsInternalFields(t *testing.T) {
	t.Parallel()

	req := Request{
		ConversationID: "conv-1",
		Messages: []domain.Message{{
			ID:             "msg-1",
			ConversationID: "conv-1",
			Role:           domain.RoleUser,
			CreatedAt:      time.Now(),
			Content: []domain.Block{{
				Index: 7,
				Type:  domain.BlockTypeText,
				Text:  &domain.TextBlock{Content: "hello"},
			}},
		}},
		Tools: []Tool{{
			Name:        "query",
			Description: "Run SQL",
			InputSchema: NewObjectSchema(map[string]Property{"sql": {Type: "string"}}, []string{"sql"}),
		}},
	}

	wireReq, err := toWireRequest(req)
	if err != nil {
		t.Fatalf("toWireRequest() error = %v", err)
	}
	data, err := json.Marshal(wireReq)
	if err != nil {
		t.Fatalf("json.Marshal(wireReq) error = %v", err)
	}
	payload := string(data)

	forbidden := []string{"\"conversation_id\"", "\"id\"", "\"created_at\"", "\"index\""}
	for _, key := range forbidden {
		if strings.Contains(payload, key) {
			t.Fatalf("payload contains forbidden field %s: %s", key, payload)
		}
	}
	if !strings.Contains(payload, `"messages"`) || !strings.Contains(payload, `"content"`) {
		t.Fatalf("payload missing expected fields: %s", payload)
	}
}

func TestToWireRequest_RejectsInvalidDomainBlock(t *testing.T) {
	t.Parallel()

	_, err := toWireRequest(Request{
		Messages: []domain.Message{{
			Role: domain.RoleUser,
			Content: []domain.Block{{
				Type: domain.BlockTypeText,
				Text: nil,
			}},
		}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
