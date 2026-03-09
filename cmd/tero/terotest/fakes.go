package terotest

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/auth/keyring"
)

type Services struct {
	APIOrigin       string
	ChatOrigin      string
	PowerSyncOrigin string
	AccessToken     string
	RefreshToken    string

	api       *httptest.Server
	chat      *httptest.Server
	powersync *httptest.Server
}

func StartFakeServices(t testing.TB) *Services {
	t.Helper()

	accessToken := validAccessToken(t, 24*time.Hour)
	refreshToken := "refresh-token"

	services := &Services{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	services.api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+accessToken {
			http.Error(w, `{"errors":[{"message":"Unauthorized"}]}`, http.StatusUnauthorized)
			return
		}

		var payload struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch {
		case strings.Contains(payload.Query, "ListOrganizations"):
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"organizations": map[string]any{
						"edges":      []any{},
						"totalCount": 0,
					},
				},
			})
		default:
			http.Error(w, `{"errors":[{"message":"Unexpected operation"}]}`, http.StatusBadRequest)
		}
	}))
	services.chat = httptest.NewServer(http.NotFoundHandler())
	services.powersync = httptest.NewServer(http.NotFoundHandler())

	services.APIOrigin = services.api.URL
	services.ChatOrigin = services.chat.URL
	services.PowerSyncOrigin = services.powersync.URL

	t.Cleanup(func() {
		services.api.Close()
		services.chat.Close()
		services.powersync.Close()
	})

	return services
}

func SeedTokens(t testing.TB, homeDir, env, accessToken, refreshToken string) {
	t.Helper()

	t.Setenv(keyring.EnvBackend, keyring.BackendFile)
	t.Setenv(keyring.EnvPath, SecretStorePath(homeDir, env))

	store, err := keyring.NewStore(env)
	if err != nil {
		t.Fatalf("create secret store: %v", err)
	}
	if err := store.Set(keyring.KeyAccessToken, accessToken); err != nil {
		t.Fatalf("seed access token: %v", err)
	}
	if err := store.Set(keyring.KeyRefreshToken, refreshToken); err != nil {
		t.Fatalf("seed refresh token: %v", err)
	}
}

func SecretStorePath(homeDir, env string) string {
	return filepath.Join(homeDir, ".tero", "environments", env, "secrets.json")
}

func validAccessToken(t testing.TB, ttl time.Duration) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, err := json.Marshal(map[string]int64{
		"exp": time.Now().Add(ttl).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal access token payload: %v", err)
	}
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
