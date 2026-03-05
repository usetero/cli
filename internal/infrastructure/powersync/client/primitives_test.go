package client

import "testing"

func TestWriteCheckpointParseInt(t *testing.T) {
	t.Parallel()

	v, err := WriteCheckpoint("42").ParseInt()
	if err != nil {
		t.Fatalf("ParseInt() error = %v", err)
	}
	if v != 42 {
		t.Fatalf("v = %d", v)
	}

	if _, err := WriteCheckpoint("nope").ParseInt(); err == nil {
		t.Fatal("expected parse error")
	}
}
