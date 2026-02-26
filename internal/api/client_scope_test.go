package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/usetero/cli/internal/auth/authtest"
	"github.com/usetero/cli/internal/domain"
)

func TestClient_WithAccountID_RemainsStableDuringConcurrentBaseSwitches(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			http.Error(w, "bad auth header", http.StatusUnauthorized)
			return
		}
		if got := r.Header.Get("X-Account-ID"); got != "acc-scoped" {
			http.Error(w, "bad account scope", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"ok": true,
			},
		})
	}))
	t.Cleanup(srv.Close)

	auth := &authtest.MockAuth{
		GetAccessTokenFunc: func(ctx context.Context) (string, error) {
			return "token", nil
		},
	}

	base := NewClient(srv.URL, auth)
	base.SetAccountID(domain.AccountID("acc-base"))
	scoped := base.WithAccountID(domain.AccountID("acc-scoped"))

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				base.SetAccountID(domain.AccountID("acc-a"))
				base.SetAccountID(domain.AccountID("acc-b"))
			}
		}
	}()

	for i := 0; i < 50; i++ {
		if _, err := scoped.RawQuery(context.Background(), "query { ok }", nil); err != nil {
			close(stop)
			<-done
			t.Fatalf("RawQuery() error at iter %d: %v", i, err)
		}
	}

	close(stop)
	<-done
}
