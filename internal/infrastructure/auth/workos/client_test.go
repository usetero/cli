package workos

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestStartDeviceAuthorization(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user_management/authorize/device" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"uc","verification_uri":"https://verify","verification_uri_complete":"https://verify?code=uc","expires_in":600,"interval":5}`))
	}))
	defer srv.Close()

	c, err := NewClient("client_1", []string{"aud_1"}, WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	flow, err := c.StartDeviceAuthorization(context.Background())
	if err != nil {
		t.Fatalf("start device auth: %v", err)
	}
	if flow.DeviceCode != "dc" || flow.UserCode != "uc" {
		t.Fatalf("unexpected flow: %+v", flow)
	}
}

func TestPollDeviceAuthorization_PendingError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"authorization_pending","error_description":"pending"}`))
	}))
	defer srv.Close()

	c, err := NewClient("client_1", []string{"aud_1"}, WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = c.PollDeviceAuthorization(context.Background(), "dc")
	if !errors.Is(err, ErrAuthorizationPending) {
		t.Fatalf("expected pending error, got: %v", err)
	}
}

func TestRefreshToken_SendsOrganizationAndAudience(t *testing.T) {
	t.Parallel()

	gotForm := url.Values{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user_management/authenticate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		gotForm = vals
		_, _ = w.Write([]byte(`{"access_token":"a2","refresh_token":"r2"}`))
	}))
	defer srv.Close()

	c, err := NewClient("client_1", []string{"aud_1", "aud_2"}, WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	tokens, err := c.RefreshToken(context.Background(), "refresh_1", "org_1")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if tokens.AccessToken != "a2" || tokens.RefreshToken != "r2" {
		t.Fatalf("unexpected tokens: %+v", tokens)
	}
	if gotForm.Get("organization_id") != "org_1" {
		t.Fatalf("organization_id missing: %v", gotForm)
	}
	if gotForm.Get("grant_type") != "refresh_token" {
		t.Fatalf("grant_type missing: %v", gotForm)
	}
	if len(gotForm["audience"]) != 2 {
		t.Fatalf("audience list missing: %v", gotForm["audience"])
	}
}
