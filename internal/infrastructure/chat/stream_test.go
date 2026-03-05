package chat

import (
	"strings"
	"testing"
)

func TestReadSSE(t *testing.T) {
	stream := strings.Join([]string{
		": keepalive",
		"data: {\"chat_stream_version\":\"v2\",\"conversation_id\":\"c\",\"turn_id\":\"t\",\"seq\":1,\"type\":\"text_delta\",\"text\":{\"content\":\"hi\"}}",
		"data: [DONE]",
	}, "\n")
	var events []Event
	if err := readSSE(strings.NewReader(stream), func(e Event) error { events = append(events, e); return nil }); err != nil {
		t.Fatalf("readSSE: %v", err)
	}
	if len(events) != 2 || events[0].TextContent != "hi" || !events[1].Done {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestDecodeEventBadVersion(t *testing.T) {
	_, err := decodeEvent([]byte(`{"chat_stream_version":"v1","type":"text_delta","text":{"content":"x"}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
