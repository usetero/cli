package chat

import (
	"fmt"
	"strings"
	"testing"
)

func FuzzDecodeEventData(f *testing.F) {
	f.Add([]byte(`{"chat_stream_version":"v2","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}`))
	f.Add([]byte(`{"chat_stream_version":"v2","type":"tool_use","tool_use":{"id":"tool-1","name":"query"}}`))
	f.Add([]byte(`{"chat_stream_version":"v2","error":"internal error"}`))
	f.Add([]byte(`{"chat_stream_version":"v2","type":"unknown"}`))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		e, err := decodeEventData(data)
		if err == nil && e.Type == "" {
			t.Fatalf("decodeEventData returned nil error with empty type for %q", string(data))
		}
	})
}

func FuzzReducerApplySequence(f *testing.F) {
	f.Add([]byte{0, 1, 6, 8})
	f.Add([]byte{0, 3, 4, 5, 6, 8})
	f.Add([]byte{0, 3, 3, 4, 4, 5, 5, 6, 8})
	f.Add([]byte{1, 8})

	f.Fuzz(func(t *testing.T, ops []byte) {
		r := newReducer("conv-fuzz")
		for i, op := range ops {
			seq := i + 1
			if op%16 == 0 && i > 0 {
				seq = i // intentionally non-monotonic sometimes
			}

			toolID := fmt.Sprintf("tool-%d", op%3)
			var e event
			switch op % 9 {
			case 0:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}}
			case 1:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeTextDelta, Text: &textContent{Content: strPtr(strings.Repeat("x", int(op%5)))}}
			case 2:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeThinkingDelta, Thinking: &textContent{Content: strPtr(strings.Repeat("t", int(op%5)))}}
			case 3:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: toolID, Name: "query"}}
			case 4:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeToolInputDelta, ToolUseID: toolID, ToolInputDelta: `{"k":1}`}
			case 5:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeContentBlockStop, ToolUseID: toolID}
			case 6:
				stopReason := "end_turn"
				if op%2 == 0 {
					stopReason = "tool_use"
				}
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: stopReason, InputTokens: intPtr(1), OutputTokens: intPtr(1)}}
			case 7:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Type: EventTypeMetadataUpdate, Metadata: &metadata{Title: "fuzz"}}
			default:
				e = event{ConversationID: "conv-fuzz", TurnID: "turn-fuzz", Seq: seq, Done: true}
			}

			_, _ = r.apply(e)
			if op%11 == 0 {
				_ = r.abortSnapshot("fuzz abort")
			}
		}
	})
}
