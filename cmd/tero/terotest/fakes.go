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

type FakeServicesOption func(*fakeControlPlane)

type FakeOrganization struct {
	ID       string
	Name     string
	Accounts []FakeAccount
}

type FakeAccount struct {
	ID            string
	Name          string
	Workspaces    []FakeWorkspace
	Datadog       *FakeDatadogAccount
	CreatedAtUnix time.Time
}

type FakeWorkspace struct {
	ID   string
	Name string
}

type FakeDatadogAccount struct {
	ID     string
	Name   string
	Site   string
	Status FakeDatadogStatus
}

type FakeDatadogStatus struct {
	ReadyForUse    bool
	EventCount     int
	AnalyzedCount  int
	ServiceCount   int
	ActiveServices int
}

type fakeControlPlane struct {
	organizations []FakeOrganization
}

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

func WithFakeOrganizations(orgs []FakeOrganization) FakeServicesOption {
	return func(cp *fakeControlPlane) {
		cp.organizations = append([]FakeOrganization(nil), orgs...)
	}
}

func StartFakeServices(t testing.TB, opts ...FakeServicesOption) *Services {
	t.Helper()

	accessToken := validAccessToken(t, 24*time.Hour)
	refreshToken := "refresh-token"
	controlPlane := fakeControlPlane{}
	for _, opt := range opts {
		opt(&controlPlane)
	}

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
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
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
						"edges":      controlPlane.organizationEdges(),
						"totalCount": len(controlPlane.organizations),
					},
				},
			})
		case strings.Contains(payload.Query, "ListAccounts"):
			orgID, _ := payload.Variables["organizationID"].(string)
			accounts := controlPlane.accountsForOrganization(orgID)
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"accounts": map[string]any{
						"edges":      accountEdges(accounts),
						"totalCount": len(accounts),
					},
				},
			})
		case strings.Contains(payload.Query, "ListWorkspaces"):
			accountID, _ := payload.Variables["accountID"].(string)
			workspaces := controlPlane.workspacesForAccount(accountID)
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"workspaces": map[string]any{
						"edges":      workspaceEdges(workspaces),
						"totalCount": len(workspaces),
					},
				},
			})
		case strings.Contains(payload.Query, "GetAccount"):
			accountID, _ := payload.Variables["id"].(string)
			account := controlPlane.accountByID(accountID)
			node := map[string]any{
				"id":             accountID,
				"datadogAccount": nil,
			}
			if account != nil && account.Datadog != nil {
				node["datadogAccount"] = map[string]any{
					"id":   account.Datadog.ID,
					"name": account.Datadog.Name,
					"site": account.Datadog.Site,
				}
			}
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"accounts": map[string]any{
						"edges": []any{
							map[string]any{"node": node},
						},
					},
				},
			})
		case strings.Contains(payload.Query, "GetDatadogAccountStatus"):
			datadogID, _ := payload.Variables["id"].(string)
			dd := controlPlane.datadogByID(datadogID)
			node := map[string]any{"id": datadogID}
			if dd != nil {
				node["status"] = map[string]any{
					"health":                        "OK",
					"readyForUse":                   dd.Status.ReadyForUse,
					"logEventCount":                 dd.Status.EventCount,
					"logEventAnalyzedCount":         dd.Status.AnalyzedCount,
					"logServiceCount":               dd.Status.ServiceCount,
					"logActiveServices":             dd.Status.ActiveServices,
					"disabledServices":              0,
					"inactiveServices":              0,
					"okServices":                    dd.Status.ActiveServices,
					"previewLogEventCount":          0,
					"effectiveLogEventCount":        0,
					"currentEventsPerHour":          nil,
					"currentBytesPerHour":           nil,
					"currentTotalUsdPerHour":        nil,
					"previewSavedEventsPerHour":     nil,
					"previewSavedBytesPerHour":      nil,
					"previewSavedTotalUsdPerHour":   nil,
					"effectiveSavedEventsPerHour":   nil,
					"effectiveSavedBytesPerHour":    nil,
					"effectiveSavedTotalUsdPerHour": nil,
					"refreshedAt":                   time.Now().UTC().Format(time.RFC3339),
				}
			}
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"datadogAccounts": map[string]any{
						"edges": []any{
							map[string]any{"node": node},
						},
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

func AccessTokenWithTTL(t testing.TB, ttl time.Duration) string {
	t.Helper()
	return validAccessToken(t, ttl)
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

func (cp fakeControlPlane) organizationEdges() []any {
	edges := make([]any, 0, len(cp.organizations))
	for _, org := range cp.organizations {
		edges = append(edges, map[string]any{
			"node": map[string]any{
				"id":                   org.ID,
				"name":                 org.Name,
				"createdAt":            time.Now().UTC().Format(time.RFC3339),
				"workosOrganizationID": "wo_" + org.ID,
			},
		})
	}
	return edges
}

func (cp fakeControlPlane) accountsForOrganization(orgID string) []FakeAccount {
	for _, org := range cp.organizations {
		if org.ID == orgID {
			return append([]FakeAccount(nil), org.Accounts...)
		}
	}
	return nil
}

func (cp fakeControlPlane) workspacesForAccount(accountID string) []FakeWorkspace {
	account := cp.accountByID(accountID)
	if account == nil {
		return nil
	}
	return append([]FakeWorkspace(nil), account.Workspaces...)
}

func (cp fakeControlPlane) accountByID(accountID string) *FakeAccount {
	for i := range cp.organizations {
		for j := range cp.organizations[i].Accounts {
			if cp.organizations[i].Accounts[j].ID == accountID {
				return &cp.organizations[i].Accounts[j]
			}
		}
	}
	return nil
}

func (cp fakeControlPlane) datadogByID(datadogID string) *FakeDatadogAccount {
	for i := range cp.organizations {
		for j := range cp.organizations[i].Accounts {
			account := &cp.organizations[i].Accounts[j]
			if account.Datadog != nil && account.Datadog.ID == datadogID {
				return account.Datadog
			}
		}
	}
	return nil
}

func accountEdges(accounts []FakeAccount) []any {
	edges := make([]any, 0, len(accounts))
	for _, account := range accounts {
		createdAt := account.CreatedAtUnix
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		edges = append(edges, map[string]any{
			"node": map[string]any{
				"id":        account.ID,
				"name":      account.Name,
				"createdAt": createdAt.Format(time.RFC3339),
			},
		})
	}
	return edges
}

func workspaceEdges(workspaces []FakeWorkspace) []any {
	edges := make([]any, 0, len(workspaces))
	for _, workspace := range workspaces {
		edges = append(edges, map[string]any{
			"node": map[string]any{
				"id":        workspace.ID,
				"name":      workspace.Name,
				"purpose":   "",
				"createdAt": time.Now().UTC().Format(time.RFC3339),
			},
		})
	}
	return edges
}
