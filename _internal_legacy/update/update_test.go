package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheck_NewerVersionAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name": "v0.2.0"}`))
	}))
	defer srv.Close()

	result, err := Check(context.Background(), "0.1.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Current != "v0.1.0" {
		t.Errorf("current = %q, want %q", result.Current, "v0.1.0")
	}
	if result.Latest != "v0.2.0" {
		t.Errorf("latest = %q, want %q", result.Latest, "v0.2.0")
	}
}

func TestCheck_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name": "v0.1.0"}`))
	}))
	defer srv.Close()

	result, err := Check(context.Background(), "v0.1.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestCheck_CurrentNewer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name": "v0.1.0"}`))
	}))
	defer srv.Close()

	result, err := Check(context.Background(), "v0.2.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestCheck_NetworkError(t *testing.T) {
	// Use a server that's immediately closed to simulate network failure.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	result, err := Check(context.Background(), "v0.1.0", srv.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

func TestCheck_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	result, err := Check(context.Background(), "v0.1.0", srv.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

func TestCheck_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	result, err := Check(context.Background(), "v0.1.0", srv.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}
