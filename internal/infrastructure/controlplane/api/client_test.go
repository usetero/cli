package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type tokenProviderFunc func(context.Context) (string, error)

func (f tokenProviderFunc) GetAccessToken(ctx context.Context) (string, error) { return f(ctx) }

func TestClientListOrganizations_SendsAuthAndMapsResponse(t *testing.T) {
	t.Parallel()

	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/graphql" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok_1" {
			t.Fatalf("authorization header mismatch: %q", got)
		}

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "ListOrganizations") {
			t.Fatalf("expected ListOrganizations query, got body=%s", string(body))
		}

		_, _ = w.Write([]byte(`{
			"data": {
				"organizations": {
					"edges": [
						null,
						{
							"node": {
								"id": "org_1",
								"name": "Primary",
								"createdAt": "2026-03-04T01:02:03Z",
								"workosOrganizationID": "wo_1"
							}
						}
					],
					"totalCount": 1
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, tokenProviderFunc(func(context.Context) (string, error) {
		return "tok_1", nil
	}))

	orgs, err := client.ListOrganizations(context.Background())
	if err != nil {
		t.Fatalf("list organizations: %v", err)
	}
	if !called {
		t.Fatalf("expected request to be sent")
	}
	if len(orgs) != 1 {
		t.Fatalf("expected one mapped org, got %d", len(orgs))
	}
	if orgs[0].ID != OrganizationID("org_1") || orgs[0].Name != "Primary" || orgs[0].WorkosOrganizationID != "wo_1" {
		t.Fatalf("unexpected mapped org: %+v", orgs[0])
	}
}

func TestClientValidateDatadogAPIKey_DefaultErrorMessage(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "ValidateDatadogApiKey") {
			t.Fatalf("expected ValidateDatadogApiKey query, got body=%s", string(body))
		}
		_, _ = w.Write([]byte(`{
			"data": {
				"validateDatadogApiKey": {
					"valid": false,
					"error": null
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	ok, message, err := client.ValidateDatadogAPIKey(context.Background(), "api_key", DatadogSiteUS1)
	if err != nil {
		t.Fatalf("validate datadog api key: %v", err)
	}
	if ok {
		t.Fatalf("expected invalid key")
	}
	if message != "invalid api key" {
		t.Fatalf("unexpected error message: %q", message)
	}
}

func TestClientGetAccountDatadogAccount_ReturnsNilWhenAbsent(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "GetAccount") {
			t.Fatalf("expected GetAccount query, got body=%s", string(body))
		}
		_, _ = w.Write([]byte(`{
			"data": {
				"accounts": {
					"edges": [
						{
							"node": {
								"id": "acc_1",
								"datadogAccount": null
							}
						}
					]
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	account, err := client.GetAccountDatadogAccount(context.Background(), AccountID("acc_1"))
	if err != nil {
		t.Fatalf("get account datadog account: %v", err)
	}
	if account != nil {
		t.Fatalf("expected nil datadog account when absent, got %+v", account)
	}
}

func TestClientCreateDatadogAccountWithCredentials_Validation(t *testing.T) {
	t.Parallel()

	client := NewClient("http://example.com", nil)

	tests := []struct {
		name  string
		input CreateDatadogAccountInput
		match string
	}{
		{name: "missing account id", input: CreateDatadogAccountInput{Name: "Main", Site: DatadogSiteUS1, APIKey: "a", AppKey: "b"}, match: "account id is required"},
		{name: "missing name", input: CreateDatadogAccountInput{AccountID: "acc_1", Site: DatadogSiteUS1, APIKey: "a", AppKey: "b"}, match: "name is required"},
		{name: "missing site", input: CreateDatadogAccountInput{AccountID: "acc_1", Name: "Main", APIKey: "a", AppKey: "b"}, match: "datadog site is required"},
		{name: "missing api key", input: CreateDatadogAccountInput{AccountID: "acc_1", Name: "Main", Site: DatadogSiteUS1, AppKey: "b"}, match: "api key is required"},
		{name: "missing app key", input: CreateDatadogAccountInput{AccountID: "acc_1", Name: "Main", Site: DatadogSiteUS1, APIKey: "a"}, match: "app key is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateDatadogAccountWithCredentials(context.Background(), tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.match) {
				t.Fatalf("expected error containing %q, got %v", tt.match, err)
			}
		})
	}
}

func TestClientListAccounts_PropagatesTokenProviderErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("token unavailable")
	client := NewClient("http://example.com", tokenProviderFunc(func(context.Context) (string, error) {
		return "", wantErr
	}))
	_, err := client.ListAccounts(context.Background(), OrganizationID("org_1"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected token provider error, got %v", err)
	}
}

func TestAuthTransport_SetsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	var gotAuth string
	var gotAccount string
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotAuth = req.Header.Get("Authorization")
		gotAccount = req.Header.Get("X-Account-ID")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})

	tr := &authTransport{token: "abc123", accountID: "acc_123", base: rt}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if gotAuth != "Bearer abc123" {
		t.Fatalf("authorization header mismatch: %q", gotAuth)
	}
	if gotAccount != "acc_123" {
		t.Fatalf("account header mismatch: %q", gotAccount)
	}
}

func TestClientListWorkspaces_SetsAccountHeader(t *testing.T) {
	t.Parallel()

	var gotAccount string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccount = r.Header.Get("X-Account-ID")
		_, _ = w.Write([]byte(`{
			"data": {
				"workspaces": {
					"edges": [],
					"totalCount": 0
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	if _, err := client.ListWorkspaces(context.Background(), AccountID("acc_1")); err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if gotAccount != "acc_1" {
		t.Fatalf("expected X-Account-ID header acc_1, got %q", gotAccount)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestClientListAccounts_MapsCreatedAt(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		_, _ = w.Write([]byte(`{
			"data": {
				"accounts": {
					"edges": [
						{
							"node": {
								"id": "acc_1",
								"name": "Primary",
								"createdAt": "2026-03-04T10:00:00Z"
							}
						}
					],
					"totalCount": 1
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	accounts, err := client.ListAccounts(context.Background(), OrganizationID("org_1"))
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected one account, got %d", len(accounts))
	}
	if !accounts[0].CreatedAt.Equal(time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected createdAt mapping: %s", accounts[0].CreatedAt)
	}
}

func TestClientDeleteOrganization_SendsConfirmedMutation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		payload := string(body)
		if !strings.Contains(payload, "DeleteOrganization") {
			t.Fatalf("expected DeleteOrganization mutation, got body=%s", payload)
		}
		if !strings.Contains(payload, `"confirmation":"DELETE"`) {
			t.Fatalf("expected DELETE confirmation variable, got body=%s", payload)
		}
		_, _ = w.Write([]byte(`{"data":{"deleteOrganization":true}}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	if err := client.DeleteOrganization(context.Background(), "org_1"); err != nil {
		t.Fatalf("delete organization: %v", err)
	}
}

func TestClientDeleteAccountAndWorkspace_ValidationAndMutation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		payload := string(body)
		switch {
		case strings.Contains(payload, "DeleteAccount"):
			_, _ = w.Write([]byte(`{"data":{"deleteAccount":true}}`))
		case strings.Contains(payload, "DeleteWorkspace"):
			_, _ = w.Write([]byte(`{"data":{"deleteWorkspace":true}}`))
		default:
			t.Fatalf("unexpected mutation body=%s", payload)
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)

	if err := client.DeleteAccount(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}
	if err := client.DeleteWorkspace(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "workspace id is required") {
		t.Fatalf("expected workspace id validation error, got %v", err)
	}
	if err := client.DeleteAccount(context.Background(), "acc_1"); err != nil {
		t.Fatalf("delete account: %v", err)
	}
	if err := client.DeleteWorkspace(context.Background(), "ws_1"); err != nil {
		t.Fatalf("delete workspace: %v", err)
	}
}

func TestNewClient_PanicsWhenOriginIncludesPath(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic")
		}
		if got := strings.TrimSpace(fmt.Sprint(r)); !strings.Contains(got, "must not include a path") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	_ = NewClient("https://api.usetero.dev/graphql", nil)
}
