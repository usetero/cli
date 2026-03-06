package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSyncStream_SuccessAndHeaders(t *testing.T) {
	t.Parallel()

	var gotAuth, gotContentType, gotAccept string
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		_, _ = io.WriteString(w, "line1\n\nline2\n")
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.SetToken(AccessToken("tok"))
	var lines []string
	err := c.SyncStream(context.Background(), &SyncStreamRequest{}, func(line []byte) error {
		lines = append(lines, string(line))
		return nil
	})
	if err != nil {
		t.Fatalf("SyncStream() error = %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/sync/stream" {
		t.Fatalf("request = %s %s, want POST /sync/stream", gotMethod, gotPath)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("auth header = %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}
	if gotAccept != "application/x-ndjson" {
		t.Fatalf("accept = %q", gotAccept)
	}
	if !reflect.DeepEqual(lines, []string{"line1", "line2"}) {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestSyncStream_HandlerErrorPropagates(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "line1\n")
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	want := fmt.Errorf("boom")
	err := c.SyncStream(context.Background(), &SyncStreamRequest{}, func([]byte) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("SyncStream() error = %v, want %v", err, want)
	}
}

func TestSyncStream_HTTPStatusClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		kind   ErrorKind
	}{
		{"auth", 401, ErrorKindAuth},
		{"transient", 503, ErrorKindTransient},
		{"rate_limit", 429, ErrorKindTransient},
		{"permanent", 400, ErrorKindPermanent},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, "err")
			}))
			defer srv.Close()

			c := NewClient(srv.URL)
			err := c.SyncStream(context.Background(), &SyncStreamRequest{}, func([]byte) error { return nil })
			var apiErr *Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %v, want *Error", err)
			}
			if apiErr.Kind != tc.kind {
				t.Fatalf("kind = %v, want %v", apiErr.Kind, tc.kind)
			}
		})
	}
}

func TestGetWriteCheckpoint(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/write-checkpoint2.json" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("client_id"); got != "cid" {
			t.Fatalf("client_id = %q", got)
		}
		_, _ = io.WriteString(w, `{"write_checkpoint":"42"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.SetToken(AccessToken("tok"))
	got, err := c.GetWriteCheckpoint(context.Background(), ClientID("cid"))
	if err != nil {
		t.Fatalf("GetWriteCheckpoint() error = %v", err)
	}
	if got != WriteCheckpoint("42") {
		t.Fatalf("checkpoint = %q", got)
	}
}

func TestGetWriteCheckpoint_StatusClassification(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "nope")
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.GetWriteCheckpoint(context.Background(), ClientID("cid"))
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *Error", err)
	}
	if apiErr.Kind != ErrorKindAuth {
		t.Fatalf("kind = %v", apiErr.Kind)
	}
}
